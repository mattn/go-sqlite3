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
