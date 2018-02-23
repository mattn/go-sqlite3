package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

extern int walHook(void *pArg, sqlite3 *db, char *zDb, int nFrame);
*/
import "C"
import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

// WalHookFunc defines the signature for a WAL callback. See also
// https://sqlite.org/c3ref/wal_hook.html.
type WalHookFunc func(unsafe.Pointer, *SQLiteConn, string, int) error

// WalHook registers a callback to be invoked each time a transaction
// is written into the write-ahead-log by the given connection.
func WalHook(conn *SQLiteConn, callback WalHookFunc, arg unsafe.Pointer) {
	db := conn.db
	handle := uintptr(unsafe.Pointer(db))
	var pointer *[0]byte
	if callback == nil {
		delete(walHooks, handle)
	} else {
		walHookMu.Lock()
		defer walHookMu.Unlock()
		walHooks[handle] = &walHookInfo{conn: conn, callback: callback}
		pointer = (*[0]byte)(unsafe.Pointer(C.walHook))
	}
	C.sqlite3_wal_hook(db, pointer, arg)
}

// WalFilename returns the path to the write-ahead file associated with the
// given connection.
func WalFilename(conn *SQLiteConn) string {
	return DatabaseFilename(conn) + "-wal"
}

// ShmFilename returns the path to the shared-memory WAL index file.
func ShmFilename(conn *SQLiteConn) string {
	return DatabaseFilename(conn) + "-shm"
}

// WalSize returns the size in bytes of the write-ahead file
// associated with the given connection. If the file does not exists,
// or can't be accessed, -1 is returned.
func WalSize(conn *SQLiteConn) int64 {
	info, err := os.Stat(WalFilename(conn))
	if err != nil {
		return -1
	}
	return info.Size()
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

// WalCheckpointV2 triggers a WAL checkpoint of the database associated
// with the given connection.  See
// https://sqlite.org/c3ref/wal_checkpoint_v2.html
func WalCheckpointV2(conn *SQLiteConn, mode WalCheckpointMode) (int, int, error) {
	var size int
	var checkpointed int
	var err error

	// Convert to C types
	db := conn.db
	zDb := C.CString("main")
	eMode := C.int(mode)
	pnLog := (*C.int)(unsafe.Pointer(&size))
	pnCkpt := (*C.int)(unsafe.Pointer(&checkpointed))

	if rc := C.sqlite3_wal_checkpoint_v2(db, zDb, eMode, pnLog, pnCkpt); rc != 0 {
		err = Error{Code: ErrNo(rc)}
	}

	return size, checkpointed, err
}

// WalPersistControl defines all valid values for the sqlite_file_control
// parameter when invoked with the SQLITE_FCNTL_PERSIST_WAL op code. See
// also https://www.sqlite.org/capi3ref.html#sqlitefcntlpersistwal
type WalPersistControl int

// WAL persist modes
const (
	WalPersistDisable = WalPersistControl(0)
	WalPersistEnable  = WalPersistControl(1)
)

// WalAutoCheckpointPragma sets the value of the WAL auto-checkpoint interval.
// See https://www.sqlite.org/pragma.html#pragma_wal_autocheckpoint
func WalAutoCheckpointPragma(conn *SQLiteConn, pages int64) error {
	if err := pragmaSetAndCheck(conn, "wal_autocheckpoint", pages); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to set wal auto checkpoint to '%d'", pages))
	}
	return nil
}

//export walHook
func walHook(pArg unsafe.Pointer, db *C.sqlite3, zDb *C.char, nFrame C.int) C.int {
	walHookMu.RLock()
	info, ok := walHooks[uintptr(unsafe.Pointer(db))]
	walHookMu.RUnlock()
	if !ok {
		panic("WAL hook not found")
	}
	if err := info.callback(pArg, info.conn, C.GoString(zDb), int(nFrame)); err != nil {
		if sqliteErr, ok := err.(Error); ok {
			return C.int(sqliteErr.Code)
		}
		return C.SQLITE_ERROR
	}
	return 0
}

type walHookInfo struct {
	conn     *SQLiteConn
	callback WalHookFunc
}

// Map C.sqlite3 pointers objects to their registered callbacks and
// connections.
var walHooks = map[uintptr]*walHookInfo{}
var walHookMu = sync.RWMutex{}
