//go:build !libsqlite3 || sqlite_serialize
// +build !libsqlite3 sqlite_serialize

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h"
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"fmt"
	"math"
	"unsafe"
)

// sqliteMaxAllocationSize matches SQLite's SQLITE_MAX_ALLOCATION_SIZE.
// See https://www.sqlite.org/security.html
const sqliteMaxAllocationSize = 2147483391

// Serialize returns a byte slice that is a serialization of the database.
//
// See https://www.sqlite.org/c3ref/serialize.html
func (c *SQLiteConn) Serialize(schema string) ([]byte, error) {
	if schema == "" {
		schema = "main"
	}
	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	var sz C.sqlite3_int64
	ptr := C.sqlite3_serialize(c.db, zSchema, &sz, 0)
	if ptr == nil {
		return nil, fmt.Errorf("serialize failed")
	}
	defer C.sqlite3_free(unsafe.Pointer(ptr))

	if sz > C.sqlite3_int64(math.MaxInt) {
		return nil, fmt.Errorf("serialized database is too large (%d bytes)", sz)
	}

	res := make([]byte, int(sz))
	copy(res, unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(sz)))
	return res, nil
}

// Deserialize causes the connection to disconnect from the current database and
// then re-open as an in-memory database based on the contents of the byte slice.
// The slice is pinned until the connection is closed; callers must not modify b
// while the connection is open.
//
// See https://www.sqlite.org/c3ref/deserialize.html
func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	if schema == "" {
		schema = "main"
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.db == nil {
		return fmt.Errorf("deserialize failed: connection closed")
	}

	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	var pData *C.uchar
	if len(b) > 0 {
		pData = (*C.uchar)(unsafe.Pointer(unsafe.SliceData(b)))
	}

	// Caller retains ownership of b; do not use SQLITE_DESERIALIZE_FREEONCLOSE,
	// which would require a buffer from sqlite3_malloc64 (limited to ~2GiB).
	rc := C.sqlite3_deserialize(c.db, zSchema, pData, C.sqlite3_int64(len(b)),
		C.sqlite3_int64(len(b)), 0)
	if rc != C.SQLITE_OK {
		return c.lastError()
	}
	c.deserializeBuf = b
	return nil
}
