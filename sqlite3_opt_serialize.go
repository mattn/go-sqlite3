//go:build !libsqlite3 || sqlite_serialize
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
	"io"
	"math"
	"reflect"
	"unsafe"
)

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

	cBuf := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(ptr)),
		Len:  int(sz),
		Cap:  int(sz),
	}))

	res := make([]byte, int(sz))
	copy(res, cBuf)
	return res, nil
}

type rawPointerReadCloser struct {
	data    unsafe.Pointer
	buf     []byte
	padding int
	sz      int
}

func (rc *rawPointerReadCloser) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if rc.padding == rc.sz {
		return 0, io.EOF
	}
	n = copy(p, rc.buf[rc.padding:])
	rc.padding += n
	return n, nil
}

func (rc *rawPointerReadCloser) Close() error {
	C.sqlite3_free(rc.data)
	return nil
}

func newRawPointerReadCloser(ptr unsafe.Pointer, sz int) *rawPointerReadCloser {
	cBuf := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(ptr),
		Len:  sz,
		Cap:  sz,
	}))
	return &rawPointerReadCloser{
		data:    ptr,
		sz:      sz,
		buf:     cBuf,
		padding: 0,
	}
}

// SerializeReader returns an io.ReadCloser of serialization of the database.
//
// See https://www.sqlite.org/c3ref/serialize.html
func (c *SQLiteConn) SerializeReader(schema string) (io.ReadCloser, error) {
	if schema == "" {
		schema = "main"
	}
	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	var sz C.sqlite3_int64
	ptr := C.sqlite3_serialize(c.db, zSchema, &sz, 0)
	if ptr == nil {
		return nil, fmt.Errorf("serialize failed: %s", C.GoString(C.sqlite3_errmsg(c.db)))
	}

	if sz > C.sqlite3_int64(math.MaxInt) {
		return nil, fmt.Errorf("serialized database is too large (%d bytes)", sz)
	}

	return newRawPointerReadCloser(unsafe.Pointer(ptr), int(sz)), nil
}

// Deserialize causes the connection to disconnect from the current database and
// then re-open as an in-memory database based on the contents of the byte slice.
//
// See https://www.sqlite.org/c3ref/deserialize.html
func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	if schema == "" {
		schema = "main"
	}
	var zSchema *C.char
	zSchema = C.CString(schema)
	defer C.free(unsafe.Pointer(zSchema))

	tmpBuf := (*C.uchar)(C.sqlite3_malloc64(C.sqlite3_uint64(len(b))))
	cBuf := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(tmpBuf)),
		Len:  len(b),
		Cap:  len(b),
	}))
	copy(cBuf, b)

	rc := C.sqlite3_deserialize(c.db, zSchema, tmpBuf, C.sqlite3_int64(len(b)),
		C.sqlite3_int64(len(b)), C.SQLITE_DESERIALIZE_FREEONCLOSE)
	if rc != C.SQLITE_OK {
		return fmt.Errorf("deserialize failed with return %v", rc)
	}
	return nil
}
