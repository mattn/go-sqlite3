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

static int
_sqlite3_bind_text(sqlite3_stmt *stmt, int n, char *p, int np) {
  return sqlite3_bind_text(stmt, n, p, np, SQLITE_TRANSIENT);
}

static int
_sqlite3_bind_blob(sqlite3_stmt *stmt, int n, void *p, int np) {
  return sqlite3_bind_blob(stmt, n, p, np, SQLITE_TRANSIENT);
}

static int
_sqlite3_step(sqlite3_stmt* stmt, long long* rowid, long long* changes)
{
  int rv = sqlite3_step(stmt);
  sqlite3* db = sqlite3_db_handle(stmt);
  *rowid = (long long) sqlite3_last_insert_rowid(db);
  *changes = (long long) sqlite3_changes(db);
  return rv;
}
*/
import "C"
import (
	"context"
	"database/sql/driver"
	"errors"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

var (
	_ driver.Stmt = (*SQLiteStmt)(nil)
)

// SQLiteStmt implement sql.Stmt.
type SQLiteStmt struct {
	mu     sync.Mutex
	c      *SQLiteConn
	s      *C.sqlite3_stmt
	t      string
	closed bool
	cls    bool
}

// Close the statement.
func (s *SQLiteStmt) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if !s.c.dbConnOpen() {
		return errors.New("sqlite statement with already closed database connection")
	}
	rv := C.sqlite3_finalize(s.s)
	s.s = nil
	if rv != C.SQLITE_OK {
		return s.c.lastError()
	}
	runtime.SetFinalizer(s, nil)
	return nil
}

// NumInput return a number of parameters.
func (s *SQLiteStmt) NumInput() int {
	return int(C.sqlite3_bind_parameter_count(s.s))
}

type bindArg struct {
	n int
	v driver.Value
}

var placeHolder = []byte{0}

func (s *SQLiteStmt) bind(args []namedValue) error {
	rv := C.sqlite3_reset(s.s)
	if rv != C.SQLITE_ROW && rv != C.SQLITE_OK && rv != C.SQLITE_DONE {
		return s.c.lastError()
	}

	for i, v := range args {
		if v.Name != "" {
			cname := C.CString(":" + v.Name)
			args[i].Ordinal = int(C.sqlite3_bind_parameter_index(s.s, cname))
			C.free(unsafe.Pointer(cname))
		}
	}

	for _, arg := range args {
		n := C.int(arg.Ordinal)
		switch v := arg.Value.(type) {
		case nil:
			rv = C.sqlite3_bind_null(s.s, n)
		case string:
			if len(v) == 0 {
				rv = C._sqlite3_bind_text(s.s, n, (*C.char)(unsafe.Pointer(&placeHolder[0])), C.int(0))
			} else {
				b := []byte(v)
				rv = C._sqlite3_bind_text(s.s, n, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)))
			}
		case int64:
			rv = C.sqlite3_bind_int64(s.s, n, C.sqlite3_int64(v))
		case bool:
			if v {
				rv = C.sqlite3_bind_int(s.s, n, 1)
			} else {
				rv = C.sqlite3_bind_int(s.s, n, 0)
			}
		case float64:
			rv = C.sqlite3_bind_double(s.s, n, C.double(v))
		case []byte:
			if v == nil {
				rv = C.sqlite3_bind_null(s.s, n)
			} else {
				ln := len(v)
				if ln == 0 {
					v = placeHolder
				}
				rv = C._sqlite3_bind_blob(s.s, n, unsafe.Pointer(&v[0]), C.int(ln))
			}
		case time.Time:
			b := []byte(v.Format(SQLiteTimestampFormats[0]))
			rv = C._sqlite3_bind_text(s.s, n, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)))
		}
		if rv != C.SQLITE_OK {
			return s.c.lastError()
		}
	}
	return nil
}

// Query the statement with arguments. Return records.
func (s *SQLiteStmt) Query(args []driver.Value) (driver.Rows, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return s.query(context.Background(), list)
}

func (s *SQLiteStmt) query(ctx context.Context, args []namedValue) (driver.Rows, error) {
	if err := s.bind(args); err != nil {
		return nil, err
	}

	rows := &SQLiteRows{
		s:        s,
		nc:       int(C.sqlite3_column_count(s.s)),
		cols:     nil,
		decltype: nil,
		cls:      s.cls,
		closed:   false,
		done:     make(chan struct{}),
	}

	if ctxdone := ctx.Done(); ctxdone != nil {
		go func(db *C.sqlite3) {
			select {
			case <-ctxdone:
				select {
				case <-rows.done:
				default:
					C.sqlite3_interrupt(db)
					rows.Close()
				}
			case <-rows.done:
			}
		}(s.c.db)
	}

	return rows, nil
}

// Exec execute the statement with arguments. Return result object.
func (s *SQLiteStmt) Exec(args []driver.Value) (driver.Result, error) {
	list := make([]namedValue, len(args))
	for i, v := range args {
		list[i] = namedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return s.exec(context.Background(), list)
}

func (s *SQLiteStmt) exec(ctx context.Context, args []namedValue) (driver.Result, error) {
	if err := s.bind(args); err != nil {
		C.sqlite3_reset(s.s)
		C.sqlite3_clear_bindings(s.s)
		return nil, err
	}

	if ctxdone := ctx.Done(); ctxdone != nil {
		done := make(chan struct{})
		defer close(done)
		go func(db *C.sqlite3) {
			select {
			case <-done:
			case <-ctxdone:
				select {
				case <-done:
				default:
					C.sqlite3_interrupt(db)
				}
			}
		}(s.c.db)
	}

	var rowid, changes C.longlong
	rv := C._sqlite3_step(s.s, &rowid, &changes)
	if rv != C.SQLITE_ROW && rv != C.SQLITE_OK && rv != C.SQLITE_DONE {
		err := s.c.lastError()
		C.sqlite3_reset(s.s)
		C.sqlite3_clear_bindings(s.s)
		return nil, err
	}

	return &SQLiteResult{id: int64(rowid), changes: int64(changes)}, nil
}
