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
*/
import "C"

//export preUpdateHookTrampoline
func preUpdateHookTrampoline(handle uintptr, dbHandle uintptr, op int, db *C.char, table *C.char, oldrowid int64, newrowid int64) {
	hval := lookupHandleVal(handle)
	data := SQLitePreUpdateData{
		Conn:         hval.db,
		Op:           op,
		DatabaseName: C.GoString(db),
		TableName:    C.GoString(table),
		OldRowID:     oldrowid,
		NewRowID:     newrowid,
	}
	callback := hval.val.(func(SQLitePreUpdateData))
	callback(data)
}
