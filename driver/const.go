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
*/
import "C"

const (
	SQLiteDelete = C.SQLITE_DELETE
	SQLiteInsert = C.SQLITE_INSERT
	SQLiteUpdate = C.SQLITE_UPDATE
)

// Backward compliant
const (
	SQLITE_DELETE = SQLiteDelete
	SQLITE_INSERT = SQLiteInsert
	SQLITE_UPDATE = SQLiteUpdate
)

const (
	columnDate      string = "date"
	columnDatetime  string = "datetime"
	columnTimestamp string = "timestamp"
)

// Run-Time Limit Categories.
// See: http://www.sqlite.org/c3ref/c_limit_attached.html
const (
	SQLITE_LIMIT_LENGTH              = C.SQLITE_LIMIT_LENGTH
	SQLITE_LIMIT_SQL_LENGTH          = C.SQLITE_LIMIT_SQL_LENGTH
	SQLITE_LIMIT_COLUMN              = C.SQLITE_LIMIT_COLUMN
	SQLITE_LIMIT_EXPR_DEPTH          = C.SQLITE_LIMIT_EXPR_DEPTH
	SQLITE_LIMIT_COMPOUND_SELECT     = C.SQLITE_LIMIT_COMPOUND_SELECT
	SQLITE_LIMIT_VDBE_OP             = C.SQLITE_LIMIT_VDBE_OP
	SQLITE_LIMIT_FUNCTION_ARG        = C.SQLITE_LIMIT_FUNCTION_ARG
	SQLITE_LIMIT_ATTACHED            = C.SQLITE_LIMIT_ATTACHED
	SQLITE_LIMIT_LIKE_PATTERN_LENGTH = C.SQLITE_LIMIT_LIKE_PATTERN_LENGTH
	SQLITE_LIMIT_VARIABLE_NUMBER     = C.SQLITE_LIMIT_VARIABLE_NUMBER
	SQLITE_LIMIT_TRIGGER_DEPTH       = C.SQLITE_LIMIT_TRIGGER_DEPTH
	SQLITE_LIMIT_WORKER_THREADS      = C.SQLITE_LIMIT_WORKER_THREADS
)
