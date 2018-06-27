// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

#include <stdlib.h>
*/
import "C"
import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var (
	_ driver.Conn   = (*SQLiteConn)(nil)
	_ driver.Execer = (*SQLiteConn)(nil)
)

// SQLiteConn implement sql.Conn.
type SQLiteConn struct {
	cfg         *Config
	mu          sync.Mutex
	db          *C.sqlite3
	tz          *time.Location
	txlock      string
	funcs       []*functionInfo
	aggregators []*aggInfo
}

// PRAGMA executes a PRAGMA statement
func (c *SQLiteConn) PRAGMA(name, value string) error {
	stmt := fmt.Sprintf("PRAGMA %s = %s;", name, value)

	cs := C.CString(stmt)
	rv := C.sqlite3_exec(c.db, cs, nil, nil, nil)
	C.free(unsafe.Pointer(cs))

	if rv != C.SQLITE_OK {
		return lastError(c.db)
	}

	return nil
}

type functionInfo struct {
	f                 reflect.Value
	argConverters     []callbackArgConverter
	variadicConverter callbackArgConverter
	retConverter      callbackRetConverter
}

func (fi *functionInfo) Call(ctx *C.sqlite3_context, argv []*C.sqlite3_value) {
	args, err := callbackConvertArgs(argv, fi.argConverters, fi.variadicConverter)
	if err != nil {
		callbackError(ctx, err)
		return
	}

	ret := fi.f.Call(args)

	if len(ret) == 2 && ret[1].Interface() != nil {
		callbackError(ctx, ret[1].Interface().(error))
		return
	}

	err = fi.retConverter(ctx, ret[0])
	if err != nil {
		callbackError(ctx, err)
		return
	}
}

type aggInfo struct {
	constructor reflect.Value

	// Active aggregator objects for aggregations in flight. The
	// aggregators are indexed by a counter stored in the aggregation
	// user data space provided by sqlite.
	active map[int64]reflect.Value
	next   int64

	stepArgConverters     []callbackArgConverter
	stepVariadicConverter callbackArgConverter

	doneRetConverter callbackRetConverter
}

func (ai *aggInfo) agg(ctx *C.sqlite3_context) (int64, reflect.Value, error) {
	aggIdx := (*int64)(C.sqlite3_aggregate_context(ctx, C.int(8)))
	if *aggIdx == 0 {
		*aggIdx = ai.next
		ret := ai.constructor.Call(nil)
		if len(ret) == 2 && ret[1].Interface() != nil {
			return 0, reflect.Value{}, ret[1].Interface().(error)
		}
		if ret[0].IsNil() {
			return 0, reflect.Value{}, errors.New("aggregator constructor returned nil state")
		}
		ai.next++
		ai.active[*aggIdx] = ret[0]
	}
	return *aggIdx, ai.active[*aggIdx], nil
}

func (ai *aggInfo) Step(ctx *C.sqlite3_context, argv []*C.sqlite3_value) {
	_, agg, err := ai.agg(ctx)
	if err != nil {
		callbackError(ctx, err)
		return
	}

	args, err := callbackConvertArgs(argv, ai.stepArgConverters, ai.stepVariadicConverter)
	if err != nil {
		callbackError(ctx, err)
		return
	}

	ret := agg.MethodByName("Step").Call(args)
	if len(ret) == 1 && ret[0].Interface() != nil {
		callbackError(ctx, ret[0].Interface().(error))
		return
	}
}

func (ai *aggInfo) Done(ctx *C.sqlite3_context) {
	idx, agg, err := ai.agg(ctx)
	if err != nil {
		callbackError(ctx, err)
		return
	}
	defer func() { delete(ai.active, idx) }()

	ret := agg.MethodByName("Done").Call(nil)
	if len(ret) == 2 && ret[1].Interface() != nil {
		callbackError(ctx, ret[1].Interface().(error))
		return
	}

	err = ai.doneRetConverter(ctx, ret[0])
	if err != nil {
		callbackError(ctx, err)
		return
	}
}

type namedValue struct {
	Name    string
	Ordinal int
	Value   driver.Value
}

// Query implements Queryer.
func (c *SQLiteConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return c.query(context.Background(), query, list)
}

func (c *SQLiteConn) query(ctx context.Context, query string, args []namedValue) (driver.Rows, error) {
	start := 0
	for {
		s, err := c.prepare(ctx, query)
		if err != nil {
			return nil, err
		}
		s.(*SQLiteStmt).cls = true
		na := s.NumInput()
		if len(args) < na {
			return nil, fmt.Errorf("not enough args to execute query: want %d got %d", na, len(args))
		}
		for i := 0; i < na; i++ {
			args[i].Ordinal -= start
		}
		rows, err := s.(*SQLiteStmt).query(ctx, args[:na])
		if err != nil && err != driver.ErrSkip {
			s.Close()
			return rows, err
		}
		args = args[na:]
		start += na
		tail := s.(*SQLiteStmt).t
		if tail == "" {
			return rows, nil
		}
		rows.Close()
		s.Close()
		query = tail
	}
}

// Exec implements Execer.
func (c *SQLiteConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return c.exec(context.Background(), query, list)
}

func (c *SQLiteConn) exec(ctx context.Context, query string, args []namedValue) (driver.Result, error) {
	start := 0
	for {
		s, err := c.prepare(ctx, query)
		if err != nil {
			return nil, err
		}
		var res driver.Result
		if s.(*SQLiteStmt).s != nil {
			na := s.NumInput()
			if len(args) < na {
				s.Close()
				return nil, fmt.Errorf("not enough args to execute query: want %d got %d", na, len(args))
			}
			for i := 0; i < na; i++ {
				args[i].Ordinal -= start
			}
			res, err = s.(*SQLiteStmt).exec(ctx, args[:na])
			if err != nil && err != driver.ErrSkip {
				s.Close()
				return nil, err
			}
			args = args[na:]
			start += na
		}
		tail := s.(*SQLiteStmt).t
		s.Close()
		if tail == "" {
			return res, nil
		}
		query = tail
	}
}

// AutoCommit return which currently auto commit or not.
func (c *SQLiteConn) AutoCommit() bool {
	return int(C.sqlite3_get_autocommit(c.db)) != 0
}

// GetFilename returns the absolute path to the file containing
// the requested schema. When passed an empty string, it will
// instead use the database's default schema: "main".
// See: sqlite3_db_filename, https://www.sqlite.org/c3ref/db_filename.html
func (c *SQLiteConn) GetFilename(schemaName string) string {
	if schemaName == "" {
		schemaName = "main"
	}
	return C.GoString(C.sqlite3_db_filename(c.db, C.CString(schemaName)))
}

// Close the connection.
func (c *SQLiteConn) Close() error {
	rv := C.sqlite3_close_v2(c.db)
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	deleteHandles(c)
	c.mu.Lock()
	c.db = nil
	c.mu.Unlock()
	runtime.SetFinalizer(c, nil)
	return nil
}

func (c *SQLiteConn) dbConnOpen() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db != nil
}

// Prepare the query string. Return a new statement.
func (c *SQLiteConn) Prepare(query string) (driver.Stmt, error) {
	return c.prepare(context.Background(), query)
}

func (c *SQLiteConn) prepare(ctx context.Context, query string) (driver.Stmt, error) {
	pquery := C.CString(query)
	defer C.free(unsafe.Pointer(pquery))
	var s *C.sqlite3_stmt
	var tail *C.char
	rv := C.sqlite3_prepare_v2(c.db, pquery, -1, &s, &tail)
	if rv != C.SQLITE_OK {
		return nil, c.lastError()
	}
	var t string
	if tail != nil && *tail != '\000' {
		t = strings.TrimSpace(C.GoString(tail))
	}
	ss := &SQLiteStmt{c: c, s: s, t: t}
	runtime.SetFinalizer(ss, (*SQLiteStmt).Close)
	return ss, nil
}
