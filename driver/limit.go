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
	// SQLiteLimitLength defines the maximum size of any string or BLOB or table row, in bytes.
	SQLiteLimitLength = C.SQLITE_LIMIT_LENGTH

	// SQLiteLimitSQLLength defines the maximum length of an SQL statement, in bytes.
	SQLiteLimitSQLLength = C.SQLITE_LIMIT_SQL_LENGTH

	// SQLiteLimitColumn defines the maximum number of columns in a table definition
	// or in the result set of a SELECT or the maximum number of columns
	// in an index or in an ORDER BY or GROUP BY clause.
	SQLiteLimitColumn = C.SQLITE_LIMIT_COLUMN

	// SQLiteLimitExprDepth defines the maximum depth of the parse tree on any expression.
	SQLiteLimitExprDepth = C.SQLITE_LIMIT_EXPR_DEPTH

	// SQLiteLimitCompoundSelect defines the maximum number of terms in a compound SELECT statement.
	SQLiteLimitCompoundSelect = C.SQLITE_LIMIT_COMPOUND_SELECT

	// SQLiteLimitVDBEOp defines the maximum number of instructions
	// in a virtual machine program used to implement an SQL statement.
	// If sqlite3_prepare_v2() or the equivalent tries to allocate space
	// for more than this many opcodes in a single prepared statement,
	// an SQLITE_NOMEM error is returned.
	SQLiteLimitVDBEOp = C.SQLITE_LIMIT_VDBE_OP

	// SQLiteLimitFunctionArg defines the maximum number of arguments on a function.
	SQLiteLimitFunctionArg = C.SQLITE_LIMIT_FUNCTION_ARG

	// SQLiteLimitAttached defines the maximum number of attached databases.
	SQLiteLimitAttached = C.SQLITE_LIMIT_ATTACHED

	// SQLiteLimitLikePatternLength defines the maximum length of the pattern argument to the LIKE or GLOB operators.
	SQLiteLimitLikePatternLength = C.SQLITE_LIMIT_LIKE_PATTERN_LENGTH

	// SQLiteLimitVariableNumber defines the maximum index number of any parameter in an SQL statement.
	SQLiteLimitVariableNumber = C.SQLITE_LIMIT_VARIABLE_NUMBER

	// SQLiteLimitTriggerDepth defines the maximum depth of recursion for triggers.
	SQLiteLimitTriggerDepth = C.SQLITE_LIMIT_TRIGGER_DEPTH

	// SQLiteLimitWorkerThreads defines the maximum number
	// of auxiliary worker threads that a single prepared statement may start.
	SQLiteLimitWorkerThreads = C.SQLITE_LIMIT_WORKER_THREADS
)

const (
	// SQLITE_LIMIT_LENGTH defines the maximum size of any string or BLOB or table row, in bytes.
	//
	// Deprecated: Use SQLiteLimitLength instead.
	SQLITE_LIMIT_LENGTH = SQLiteLimitLength

	// SQLITE_LIMIT_SQL_LENGTH defines the maximum length of an SQL statement, in bytes.
	//
	// Deprecated: Use SQLiteLimitSQLLength instead.
	SQLITE_LIMIT_SQL_LENGTH = SQLiteLimitSQLLength

	// SQLITE_LIMIT_COLUMN defines the maximum number of columns in a table definition
	// or in the result set of a SELECT or the maximum number of columns
	// in an index or in an ORDER BY or GROUP BY clause.
	//
	// Deprecated: Use SQLiteLimitColumn instead.
	SQLITE_LIMIT_COLUMN = SQLiteLimitColumn

	// SQLITE_LIMIT_EXPR_DEPTH defines the maximum depth of the parse tree on any expression.
	//
	// Deprecated: Use SQLiteLimitExprDepth instead.
	SQLITE_LIMIT_EXPR_DEPTH = SQLiteLimitExprDepth

	// SQLITE_LIMIT_COMPOUND_SELECT defines the maximum number of terms in a compound SELECT statement.
	//
	// Deprecated: Use SQLiteLimitCompoundSelect instead.
	SQLITE_LIMIT_COMPOUND_SELECT = SQLiteLimitCompoundSelect

	// SQLITE_LIMIT_VDBE_OP defines the maximum number of instructions
	// in a virtual machine program used to implement an SQL statement.
	// If sqlite3_prepare_v2() or the equivalent tries to allocate space
	// for more than this many opcodes in a single prepared statement,
	// an SQLITE_NOMEM error is returned.
	//
	// Deprecated: Use SQLiteLimitVDBEOp instead.
	SQLITE_LIMIT_VDBE_OP = SQLiteLimitVDBEOp

	// SQLITE_LIMIT_FUNCTION_ARG defines the maximum number of arguments on a function.
	//
	// Deprecated: Use SQLiteLimitFunctionArg instead.
	SQLITE_LIMIT_FUNCTION_ARG = SQLiteLimitFunctionArg

	// SQLITE_LIMIT_ATTACHED defines the maximum number of attached databases.
	//
	// Deprecated: Use SQLiteLimitAttached instead.
	SQLITE_LIMIT_ATTACHED = SQLiteLimitAttached

	// SQLITE_LIMIT_LIKE_PATTERN_LENGTH defines the maximum length of the pattern argument to the LIKE or GLOB operators.
	//
	// Deprecated: Use SQLiteLimitLikePatternLength instead.
	SQLITE_LIMIT_LIKE_PATTERN_LENGTH = SQLiteLimitLikePatternLength

	// SQLITE_LIMIT_VARIABLE_NUMBER defines the maximum index number of any parameter in an SQL statement.
	//
	// Deprecated: Use SQLiteLimitVariableNumber instead.
	SQLITE_LIMIT_VARIABLE_NUMBER = SQLiteLimitVariableNumber

	// SQLITE_LIMIT_TRIGGER_DEPTH defines the maximum depth of recursion for triggers.
	//
	// Deprecated: Use SQLiteLimitTriggerDepth instead.
	SQLITE_LIMIT_TRIGGER_DEPTH = SQLiteLimitTriggerDepth

	// SQLITE_LIMIT_WORKER_THREADS defines the maximum number
	// of auxiliary worker threads that a single prepared statement may start.
	//
	// Deprecated: Use SQLiteLimitWorkerThreads instead.
	SQLITE_LIMIT_WORKER_THREADS = SQLiteLimitWorkerThreads
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
