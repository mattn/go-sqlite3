// +build !sqlite_omit_deserialize

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Serialize returns a byte slice that is a serialization of the database.
// If the database fails to serialize, a nil slice will be returned.
//
// See https://www.sqlite.org/c3ref/serialize.html
func (c *SQLiteConn) Serialize(schema string) []byte {
	if schema == "" {
		schema = "main"
	}
	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	var sz C.sqlite3_int64
	ptr := C.sqlite3_serialize(c.db, zSchema, &sz, 0)
	if ptr == nil {
		return nil
	}
	defer C.sqlite3_free(unsafe.Pointer(ptr))
	return C.GoBytes(unsafe.Pointer(ptr), C.int(sz))
}

// Deserialize causes the connection to disconnect from the current database
// and then re-open as an in-memory database based on the contents of the
// byte slice. If deserelization fails, error will contain the return code
// of the underlying SQLite API call.
//
// When this function returns, the connection is referencing database
// data in Go space, so the connection and associated database must be copied
// immediately if it is to be used further. SQLiteConn.Backup() can be used
// to perform this copy.
//
// See https://www.sqlite.org/c3ref/deserialize.html
func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	if schema == "" {
		schema = "main"
	}
	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	rc := C.sqlite3_deserialize(c.db, zSchema,
		(*C.uint8_t)(unsafe.Pointer(&b[0])),
		C.sqlite3_int64(len(b)), C.sqlite3_int64(len(b)), 0)
	if rc != 0 {
		return fmt.Errorf("deserialize failed with return %v", rc)
	}
	return nil
}
