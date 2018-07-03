// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
#cgo CFLAGS: -std=gnu99
#cgo CFLAGS: -DSQLITE_ENABLE_RTREE
#cgo CFLAGS: -DSQLITE_THREADSAFE=1
#cgo CFLAGS: -DHAVE_USLEEP=1
#cgo CFLAGS: -DSQLITE_ENABLE_FTS3
#cgo CFLAGS: -DSQLITE_ENABLE_FTS3_PARENTHESIS
#cgo CFLAGS: -DSQLITE_ENABLE_FTS4_UNICODE61
#cgo CFLAGS: -DSQLITE_TRACE_SIZE_LIMIT=15
#cgo CFLAGS: -DSQLITE_OMIT_DEPRECATED
#cgo CFLAGS: -DSQLITE_DISABLE_INTRINSIC
#cgo CFLAGS: -DSQLITE_DEFAULT_WAL_SYNCHRONOUS=1
#cgo CFLAGS: -DSQLITE_ENABLE_UPDATE_DELETE_LIMIT
#cgo CFLAGS: -Wno-deprecated-declarations
#cgo linux,!android CFLAGS: -DHAVE_PREAD64=1 -DHAVE_PWRITE64=1

#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
*/
import "C"
import (
	"database/sql"
	"database/sql/driver"
	"sync"
)

var (
	_ driver.Driver = (*SQLiteDriver)(nil)
)

func init() {
	sql.Register("sqlite3", &SQLiteDriver{})
}

// SQLiteDriver implement sql.Driver.
type SQLiteDriver struct {
	mu          sync.Mutex
	Config      *Config
	Extensions  []string
	ConnectHook func(*SQLiteConn) error
}

// Open database and return a new connection.
func (d *SQLiteDriver) Open(dsn string) (driver.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	cfg, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// Configure Extensions
	cfg.Extensions = d.Extensions

	// Configure ConnectHook
	cfg.ConnectHook = d.ConnectHook

	// Set Configuration
	d.Config = cfg

	return cfg.createConnection()
}
