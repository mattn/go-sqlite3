// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
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
#cgo CFLAGS: -std=gnu99

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

static int
_sqlite3_open_v2(const char *filename, sqlite3 **ppDb, int flags, const char *zVfs) {
#ifdef SQLITE_OPEN_URI
  return sqlite3_open_v2(filename, ppDb, flags | SQLITE_OPEN_URI, zVfs);
#else
  return sqlite3_open_v2(filename, ppDb, flags, zVfs);
#endif
}

static int
_sqlite3_bind_text(sqlite3_stmt *stmt, int n, char *p, int np) {
  return sqlite3_bind_text(stmt, n, p, np, SQLITE_TRANSIENT);
}

static int
_sqlite3_bind_blob(sqlite3_stmt *stmt, int n, void *p, int np) {
  return sqlite3_bind_blob(stmt, n, p, np, SQLITE_TRANSIENT);
}

#include <stdio.h>
#include <stdint.h>

static int
_sqlite3_exec(sqlite3* db, const char* pcmd, long long* rowid, long long* changes)
{
  int rv = sqlite3_exec(db, pcmd, 0, 0, 0);
  *rowid = (long long) sqlite3_last_insert_rowid(db);
  *changes = (long long) sqlite3_changes(db);
  return rv;
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

void _sqlite3_result_text(sqlite3_context* ctx, const char* s) {
  sqlite3_result_text(ctx, s, -1, &free);
}

void _sqlite3_result_blob(sqlite3_context* ctx, const void* b, int l) {
  sqlite3_result_blob(ctx, b, l, SQLITE_TRANSIENT);
}


int _sqlite3_create_function(
  sqlite3 *db,
  const char *zFunctionName,
  int nArg,
  int eTextRep,
  uintptr_t pApp,
  void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
  void (*xStep)(sqlite3_context*,int,sqlite3_value**),
  void (*xFinal)(sqlite3_context*)
) {
  return sqlite3_create_function(db, zFunctionName, nArg, eTextRep, (void*) pApp, xFunc, xStep, xFinal);
}

void callbackTrampoline(sqlite3_context*, int, sqlite3_value**);
void stepTrampoline(sqlite3_context*, int, sqlite3_value**);
void doneTrampoline(sqlite3_context*);

int compareTrampoline(void*, int, char*, int, char*);
int commitHookTrampoline(void*);
void rollbackHookTrampoline(void*);
void updateHookTrampoline(void*, int, char*, char*, sqlite3_int64);

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
