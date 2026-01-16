// Copyright (C) 2022 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h"
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"unsafe"
)

// SQLiteBlob implements the SQLite Blob I/O interface.
type SQLiteBlob struct {
	conn   *SQLiteConn
	blob   *C.sqlite3_blob
	size   int
	offset int
}

// Blob opens a blob.
//
// See https://www.sqlite.org/c3ref/blob_open.html for usage.
//
// Should only be used with conn.Raw.
func (conn *SQLiteConn) Blob(database, table, column string, rowid int64, flags int) (*SQLiteBlob, error) {
	databaseptr := C.CString(database)
	defer C.free(unsafe.Pointer(databaseptr))

	tableptr := C.CString(table)
	defer C.free(unsafe.Pointer(tableptr))

	columnptr := C.CString(column)
	defer C.free(unsafe.Pointer(columnptr))

	var blob *C.sqlite3_blob
	ret := C.sqlite3_blob_open(conn.db, databaseptr, tableptr, columnptr, C.longlong(rowid), C.int(flags), &blob)

	if ret != C.SQLITE_OK {
		return nil, conn.lastError()
	}

	size := int(C.sqlite3_blob_bytes(blob))
	bb := &SQLiteBlob{conn: conn, blob: blob, size: size, offset: 0}

	runtime.SetFinalizer(bb, (*SQLiteBlob).Close)

	return bb, nil
}

// Read implements the io.Reader interface.
func (s *SQLiteBlob) Read(b []byte) (n int, err error) {
	if s.offset >= s.size {
		return 0, io.EOF
	}

	if len(b) == 0 {
		return 0, nil
	}

	n = s.size - s.offset
	if len(b) < n {
		n = len(b)
	}

	p := &b[0]
	ret := C.sqlite3_blob_read(s.blob, unsafe.Pointer(p), C.int(n), C.int(s.offset))
	if ret != C.SQLITE_OK {
		return 0, s.conn.lastError()
	}

	s.offset += n

	return n, nil
}

// Write implements the io.Writer interface.
func (s *SQLiteBlob) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}

	if s.offset >= s.size {
		return 0, fmt.Errorf("sqlite3.SQLiteBlob.Write: insufficient space in %d-byte blob", s.size)
	}

	n = s.size - s.offset
	if len(b) < n {
		n = len(b)
	}

	if n != len(b) {
		return 0, fmt.Errorf("sqlite3.SQLiteBlob.Write: insufficient space in %d-byte blob", s.size)
	}

	p := &b[0]
	ret := C.sqlite3_blob_write(s.blob, unsafe.Pointer(p), C.int(n), C.int(s.offset))
	if ret != C.SQLITE_OK {
		return 0, s.conn.lastError()
	}

	s.offset += n

	return n, nil
}

// Seek implements the io.Seeker interface.
func (s *SQLiteBlob) Seek(offset int64, whence int) (int64, error) {
	if offset > math.MaxInt32 {
		return 0, fmt.Errorf("sqlite3.SQLiteBlob.Seek: invalid offset %d", offset)
	}

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(s.offset) + offset
	case io.SeekEnd:
		abs = int64(s.size) + offset
	default:
		return 0, fmt.Errorf("sqlite3.SQLiteBlob.Seek: invalid whence %d", whence)
	}

	if abs < 0 {
		return 0, errors.New("sqlite.SQLiteBlob.Seek: negative position")
	}

	if abs > math.MaxInt32 || abs > int64(s.size) {
		return 0, errors.New("sqlite3.SQLiteBlob.Seek: overflow position")
	}

	s.offset = int(abs)

	return abs, nil
}

// Size returns the size of the blob.
func (s *SQLiteBlob) Size() int {
	return s.size
}

// Close implements the io.Closer interface.
func (s *SQLiteBlob) Close() error {
	ret := C.sqlite3_blob_close(s.blob)

	s.blob = nil
	runtime.SetFinalizer(s, nil)

	if ret != C.SQLITE_OK {
		return s.conn.lastError()
	}

	return nil
}
