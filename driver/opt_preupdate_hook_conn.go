// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build !libsqlite3
// +build sqlite_preupdate_hook

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_PREUPDATE_HOOK
#cgo LDFLAGS: -lm

#include <sqlite3-binding.h>
#include <stdlib.h>

void preUpdateHookTrampoline(void*, sqlite3 *, int, char *, char *, sqlite3_int64, sqlite3_int64);
*/
import "C"
import (
	"unsafe"
)

// RegisterPreUpdateHook sets the pre-update hook for a connection.
//
// The callback is passed a SQLitePreUpdateData struct with the data for
// the update, as well as methods for fetching copies of impacted data.
//
// If there is an existing update hook for this connection, it will be
// removed. If callback is nil the existing hook (if any) will be removed
// without creating a new one.
func (c *SQLiteConn) RegisterPreUpdateHook(callback func(SQLitePreUpdateData)) {
	if callback == nil {
		C.sqlite3_preupdate_hook(c.db, nil, nil)
	} else {
		C.sqlite3_preupdate_hook(c.db, (*[0]byte)(unsafe.Pointer(C.preUpdateHookTrampoline)), unsafe.Pointer(newHandle(c, callback)))
	}
}
