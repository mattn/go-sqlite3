package sqlite3test

/*
#cgo CFLAGS: -DUSE_LIBSQLITE3
#cgo linux LDFLAGS: -lsqlite3
#cgo darwin LDFLAGS: -L/usr/local/opt/sqlite/lib -lsqlite3
#cgo darwin CFLAGS: -I/usr/local/opt/sqlite/include
#cgo openbsd LDFLAGS: -lsqlite3
#cgo solaris LDFLAGS: -lsqlite3
#include <stdlib.h>
#include <sqlite3.h>

static void one(sqlite3_context* ctx, int n, sqlite3_value** vals) {
    sqlite3_result_int(ctx, 1);
}

static inline int _create_function(sqlite3* c) {
    return sqlite3_create_function(c, "one", 0, SQLITE_UTF8|SQLITE_DETERMINISTIC, NULL, one, NULL, NULL);
}
*/
import "C"
import (
	sqlite3 "github.com/mattn/go-sqlite3"
	"unsafe"
)

func RegisterFunction(conn *sqlite3.SQLiteConn) error {
	return conn.Raw(func(raw unsafe.Pointer) error {
		rawConn := (*C.sqlite3)(raw)
		if ret := C._create_function(rawConn); ret != C.SQLITE_OK {
			return sqlite3.ErrNo(ret)
		}
		return nil
	})
}
