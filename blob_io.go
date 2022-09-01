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
	"io"
	"runtime"
	"unsafe"
)

// SQLiteBlob implements the SQLite Blob I/O interface.
type SQLiteBlob struct {
	conn *SQLiteConn
	blob *C.sqlite3_blob
	size int
	offs int
}

// Blob opens a blob.
//
// The flag parameter is ignored.
func (conn *SQLiteConn) Blob(database, table, column string, rowid int64, flags int) (*SQLiteBlob, error) {
	databaseptr := C.CString(database)
	defer C.free(unsafe.Pointer(databaseptr))

	tableptr := C.CString(table)
	defer C.free(unsafe.Pointer(tableptr))

	columnptr := C.CString(column)
	defer C.free(unsafe.Pointer(columnptr))

	var blob *C.sqlite3_blob
	ret := C.sqlite3_blob_open(conn.db, databaseptr, tableptr, columnptr, C.longlong(rowid), C.int(flags), &blob)

	if ret == C.SQLITE_OK {
		size := int(C.sqlite3_blob_bytes(blob))
		bb := &SQLiteBlob{conn, blob, size, 0}

		runtime.SetFinalizer(bb, (*SQLiteBlob).Close)

		return bb, nil
	}

	return nil, conn.lastError()
}

// Read implements the io.Reader interface.
func (s *SQLiteBlob) Read(b []byte) (n int, err error) {
	if s.offs >= s.size {
		return 0, io.EOF
	}

	n = s.size - s.offs
	if len(b) < n {
		n = len(b)
	}

	p := &b[0]
	ret := C.sqlite3_blob_read(s.blob, unsafe.Pointer(p), C.int(n), C.int(s.offs))
	if ret != C.SQLITE_OK {
		return 0, s.conn.lastError()
	}

	s.offs += n

	return n, nil
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
