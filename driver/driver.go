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
#include <stdlib.h>
#include <string.h>

#ifdef __CYGWIN__
# include <errno.h>
#endif

#ifndef SQLITE_OPEN_READWRITE
# define SQLITE_OPEN_READWRITE 0
#endif

#ifndef SQLITE_OPEN_FULLMUTEX
# define SQLITE_OPEN_FULLMUTEX 0
#endif

#ifndef SQLITE_DETERMINISTIC
# define SQLITE_DETERMINISTIC 0
#endif
*/
import "C"
import (
	"database/sql"
	"database/sql/driver"
)

var (
	_ driver.Driver = (*SQLiteDriver)(nil)
)

func init() {
	sql.Register("sqlite3", &SQLiteDriver{})
}

// SQLiteDriver implement sql.Driver.
type SQLiteDriver struct {
	Extensions  []string
	ConnectHook func(*SQLiteConn) error
}

// Open database and return a new connection.
func (d *SQLiteDriver) Open(dsn string) (driver.Conn, error) {
	return nil, nil
}
