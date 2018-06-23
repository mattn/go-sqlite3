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
#include <string.h>

#ifdef SQLITE_LIMIT_WORKER_THREADS
# define _SQLITE_HAS_LIMIT
# define SQLITE_LIMIT_LENGTH                    0
# define SQLITE_LIMIT_SQL_LENGTH                1
# define SQLITE_LIMIT_COLUMN                    2
# define SQLITE_LIMIT_EXPR_DEPTH                3
# define SQLITE_LIMIT_COMPOUND_SELECT           4
# define SQLITE_LIMIT_VDBE_OP                   5
# define SQLITE_LIMIT_FUNCTION_ARG              6
# define SQLITE_LIMIT_ATTACHED                  7
# define SQLITE_LIMIT_LIKE_PATTERN_LENGTH       8
# define SQLITE_LIMIT_VARIABLE_NUMBER           9
# define SQLITE_LIMIT_TRIGGER_DEPTH            10
# define SQLITE_LIMIT_WORKER_THREADS           11
# else
# define SQLITE_LIMIT_WORKER_THREADS           11
#endif

static int _sqlite3_limit(sqlite3* db, int limitId, int newLimit) {
#ifndef _SQLITE_HAS_LIMIT
  return -1;
#else
  return sqlite3_limit(db, limitId, newLimit);
#endif
}
*/
import "C"

// Run-Time Limit Categories.
// See: http://www.sqlite.org/c3ref/c_limit_attached.html
const (
	// SQLITE_LIMIT_LENGTH defines the maximum size of any string or BLOB or table row, in bytes.
	SQLITE_LIMIT_LENGTH = C.SQLITE_LIMIT_LENGTH

	// SQLITE_LIMIT_SQL_LENGTH defines the maximum length of an SQL statement, in bytes.
	SQLITE_LIMIT_SQL_LENGTH = C.SQLITE_LIMIT_SQL_LENGTH

	// SQLITE_LIMIT_COLUMN defines the maximum number of columns in a table definition
	// or in the result set of a SELECT or the maximum number of columns
	// in an index or in an ORDER BY or GROUP BY clause.
	SQLITE_LIMIT_COLUMN = C.SQLITE_LIMIT_COLUMN

	// SQLITE_LIMIT_EXPR_DEPTH defines the maximum depth of the parse tree on any expression.
	SQLITE_LIMIT_EXPR_DEPTH = C.SQLITE_LIMIT_EXPR_DEPTH

	// SQLITE_LIMIT_COMPOUND_SELECT defines the maximum number of terms in a compound SELECT statement.
	SQLITE_LIMIT_COMPOUND_SELECT = C.SQLITE_LIMIT_COMPOUND_SELECT

	// SQLITE_LIMIT_VDBE_OP defines the maximum number of instructions
	// in a virtual machine program used to implement an SQL statement.
	// If sqlite3_prepare_v2() or the equivalent tries to allocate space
	// for more than this many opcodes in a single prepared statement,
	// an SQLITE_NOMEM error is returned.
	SQLITE_LIMIT_VDBE_OP = C.SQLITE_LIMIT_VDBE_OP

	// SQLITE_LIMIT_FUNCTION_ARG defines the maximum number of arguments on a function.
	SQLITE_LIMIT_FUNCTION_ARG = C.SQLITE_LIMIT_FUNCTION_ARG

	// SQLITE_LIMIT_ATTACHED defines the maximum number of attached databases.
	SQLITE_LIMIT_ATTACHED = C.SQLITE_LIMIT_ATTACHED

	// SQLITE_LIMIT_LIKE_PATTERN_LENGTH defines the maximum length of the pattern argument to the LIKE or GLOB operators.
	SQLITE_LIMIT_LIKE_PATTERN_LENGTH = C.SQLITE_LIMIT_LIKE_PATTERN_LENGTH

	// SQLITE_LIMIT_VARIABLE_NUMBER defines the maximum index number of any parameter in an SQL statement.
	SQLITE_LIMIT_VARIABLE_NUMBER = C.SQLITE_LIMIT_VARIABLE_NUMBER

	// SQLITE_LIMIT_TRIGGER_DEPTH defines the maximum depth of recursion for triggers.
	SQLITE_LIMIT_TRIGGER_DEPTH = C.SQLITE_LIMIT_TRIGGER_DEPTH

	// SQLITE_LIMIT_WORKER_THREADS defines the maximum number
	// of auxiliary worker threads that a single prepared statement may start.
	SQLITE_LIMIT_WORKER_THREADS = C.SQLITE_LIMIT_WORKER_THREADS
)

// GetLimit returns the current value of a run-time limit.
// See: sqlite3_limit, http://www.sqlite.org/c3ref/limit.html
func (c *SQLiteConn) GetLimit(id int) int {
	return int(C._sqlite3_limit(c.db, C.int(id), -1))
}

// SetLimit changes the value of a run-time limits.
// Then this method returns the prior value of the limit.
// See: sqlite3_limit, http://www.sqlite.org/c3ref/limit.html
func (c *SQLiteConn) SetLimit(id int, newVal int) int {
	return int(C._sqlite3_limit(c.db, C.int(id), C.int(newVal)))
}
