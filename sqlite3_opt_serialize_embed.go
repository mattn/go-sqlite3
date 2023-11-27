// +build go1.21
// +build !libsqlite3 sqlite_serialize

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
	"runtime"
	"unsafe"
)

// This is the reason for requiring go1.21
var embedPin = runtime.Pinner{}

// DeserializeEmbedded is different from Deserialize in that it does NOT copy
// memory and does NOT free the memory on close. In fact it pins the memory,
// and does not give you a way to unpin. Therefore this should only be used for
// an embedded, read-only database that stays for the lifetime of your process.
//
// Because of how initial DB must be opened this can only have schema of "main".
// Leaving the argument in case that ever changes (the simplest would be if SQLite
// added a mode of "rom" or similar. But it can't be fixed just in go).
func (c *SQLiteConn) DeserializeEmbedded(b []byte, schema string) error {
	// "main" is the only value `schema` can be because it must be opened with:
	//   "file::memory:?mode=ro"
	//   "file::memory:?mode=ro&cache=shared"
	//   "file::memory:?mode=ro&cache=private"
	// See https://github.com/mattn/go-sqlite3/issues/204 as this means you
	// can embed exactly one shared DB into your binary at a time.
	// Maybe there will eventually be a way to open like:
	//   "file:unique_name.db?mode=rom"
	// Then we would use the `schema` argument.
	schema = "main"
	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	// Pinning will future proof for a GC that moves memory. It is also
	// necessary now because nothing else prevents a reassignment of `b`
	// which would let GC reclaim `b`.
	rom := unsafe.Pointer(&b[0])
	embedPin.Pin(rom)

	rc := C.sqlite3_deserialize(c.db, zSchema, (*C.uchar)(rom),
		C.sqlite3_int64(len(b)), C.sqlite3_int64(len(b)),
		C.SQLITE_DESERIALIZE_READONLY)
	if rc != C.SQLITE_OK {
		return fmt.Errorf("deserialize failed with return %v", rc)
	}
	return nil
}
