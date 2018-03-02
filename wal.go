package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>

int walHookTrampoline(void*, sqlite3*, char*, int);
*/
import "C"
import (
	"unsafe"
)

// RegisterWalHook sets the wal hook for a connection.
func (c *SQLiteConn) RegisterWalHook(callback func(string, int) int) {
	if callback == nil {
		C.sqlite3_wal_hook(c.db, nil, nil)
	} else {
		handle := newHandle(c, callback)
		C.sqlite3_wal_hook(c.db, (*[0]byte)(unsafe.Pointer(C.walHookTrampoline)), unsafe.Pointer(handle))
	}
}

// WalCheckpointMode defines all valid values for the "checkpoint mode" parameter
// of the WalCheckpointV2 API. See https://sqlite.org/c3ref/wal_checkpoint_v2.html.
type WalCheckpointMode int

// WAL checkpoint modes
const (
	WalCheckpointPassive  = WalCheckpointMode(C.SQLITE_CHECKPOINT_PASSIVE)
	WalCheckpointFull     = WalCheckpointMode(C.SQLITE_CHECKPOINT_FULL)
	WalCheckpointRestart  = WalCheckpointMode(C.SQLITE_CHECKPOINT_RESTART)
	WalCheckpointTruncate = WalCheckpointMode(C.SQLITE_CHECKPOINT_TRUNCATE)
)

// WalCheckpoint triggers a WAL checkpoint on the given database attached to the
// connection. See https://sqlite.org/c3ref/wal_checkpoint_v2.html
func (c *SQLiteConn) WalCheckpoint(db string, mode WalCheckpointMode) (int, int, error) {
	var size C.int
	var ckpt C.int
	var err error

	// Convert to C types
	zDb := C.CString(db)
	defer C.free(unsafe.Pointer(zDb))

	rv := C.sqlite3_wal_checkpoint_v2(c.db, zDb, C.int(mode), &size, &ckpt)
	if rv != 0 {
		err = newError(rv)
	}

	return int(size), int(ckpt), err
}
