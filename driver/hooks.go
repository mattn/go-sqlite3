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

int commitHookTrampoline(void*);
void rollbackHookTrampoline(void*);
void updateHookTrampoline(void*, int, char*, char*, sqlite3_int64);
*/
import "C"
import "unsafe"

// RegisterCommitHook sets the commit hook for a connection.
//
// If the callback returns non-zero the transaction will become a rollback.
//
// If there is an existing commit hook for this connection, it will be
// removed. If callback is nil the existing hook (if any) will be removed
// without creating a new one.
func (c *SQLiteConn) RegisterCommitHook(callback func() int) {
	if callback == nil {
		C.sqlite3_commit_hook(c.db, nil, nil)
	} else {
		C.sqlite3_commit_hook(c.db, (*[0]byte)(C.commitHookTrampoline), unsafe.Pointer(newHandle(c, callback)))
	}
}

// RegisterRollbackHook sets the rollback hook for a connection.
//
// If there is an existing rollback hook for this connection, it will be
// removed. If callback is nil the existing hook (if any) will be removed
// without creating a new one.
func (c *SQLiteConn) RegisterRollbackHook(callback func()) {
	if callback == nil {
		C.sqlite3_rollback_hook(c.db, nil, nil)
	} else {
		C.sqlite3_rollback_hook(c.db, (*[0]byte)(C.rollbackHookTrampoline), unsafe.Pointer(newHandle(c, callback)))
	}
}

// RegisterUpdateHook sets the update hook for a connection.
//
// The parameters to the callback are the operation (one of the constants
// SQLITE_INSERT, SQLITE_DELETE, or SQLITE_UPDATE), the database name, the
// table name, and the rowid.
//
// If there is an existing update hook for this connection, it will be
// removed. If callback is nil the existing hook (if any) will be removed
// without creating a new one.
func (c *SQLiteConn) RegisterUpdateHook(callback func(int, string, string, int64)) {
	if callback == nil {
		C.sqlite3_update_hook(c.db, nil, nil)
	} else {
		C.sqlite3_update_hook(c.db, (*[0]byte)(C.updateHookTrampoline), unsafe.Pointer(newHandle(c, callback)))
	}
}
