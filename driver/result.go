// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

#include <stdlib.h>

void _sqlite3_result_text(sqlite3_context* ctx, const char* s) {
  sqlite3_result_text(ctx, s, -1, &free);
}

void _sqlite3_result_blob(sqlite3_context* ctx, const void* b, int l) {
  sqlite3_result_blob(ctx, b, l, SQLITE_TRANSIENT);
}
*/
import "C"
import (
	"database/sql/driver"
	"io"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

var (
	_ driver.Result                         = (*SQLiteResult)(nil)
	_ driver.Rows                           = (*SQLiteRows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*SQLiteRows)(nil)
	_ driver.RowsColumnTypeNullable         = (*SQLiteRows)(nil)
	_ driver.RowsColumnTypeDatabaseTypeName = (*SQLiteRows)(nil)
)

// SQLiteResult implement sql.Result.
type SQLiteResult struct {
	id      int64
	changes int64
}

// SQLiteRows implement sql.Rows.
type SQLiteRows struct {
	s        *SQLiteStmt
	nc       int
	cols     []string
	decltype []string
	cls      bool
	closed   bool
	done     chan struct{}
}

// LastInsertId teturn last inserted ID.
func (r *SQLiteResult) LastInsertId() (int64, error) {
	return r.id, nil
}

// RowsAffected return how many rows affected.
func (r *SQLiteResult) RowsAffected() (int64, error) {
	return r.changes, nil
}

// Close the rows.
func (rc *SQLiteRows) Close() error {
	rc.s.mu.Lock()
	if rc.s.closed || rc.closed {
		rc.s.mu.Unlock()
		return nil
	}
	rc.closed = true
	if rc.done != nil {
		close(rc.done)
	}
	if rc.cls {
		rc.s.mu.Unlock()
		return rc.s.Close()
	}
	rv := C.sqlite3_reset(rc.s.s)
	if rv != C.SQLITE_OK {
		rc.s.mu.Unlock()
		return rc.s.c.lastError()
	}
	rc.s.mu.Unlock()
	return nil
}

// Columns return column names.
func (rc *SQLiteRows) Columns() []string {
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()
	if rc.s.s != nil && rc.nc != len(rc.cols) {
		rc.cols = make([]string, rc.nc)
		for i := 0; i < rc.nc; i++ {
			rc.cols[i] = C.GoString(C.sqlite3_column_name(rc.s.s, C.int(i)))
		}
	}
	return rc.cols
}

func (rc *SQLiteRows) declTypes() []string {
	if rc.s.s != nil && rc.decltype == nil {
		rc.decltype = make([]string, rc.nc)
		for i := 0; i < rc.nc; i++ {
			rc.decltype[i] = strings.ToLower(C.GoString(C.sqlite3_column_decltype(rc.s.s, C.int(i))))
		}
	}
	return rc.decltype
}

// DeclTypes return column types.
func (rc *SQLiteRows) DeclTypes() []string {
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()
	return rc.declTypes()
}

// Next move cursor to next.
func (rc *SQLiteRows) Next(dest []driver.Value) error {
	if rc.s.closed {
		return io.EOF
	}
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()
	rv := C.sqlite3_step(rc.s.s)
	if rv == C.SQLITE_DONE {
		return io.EOF
	}
	if rv != C.SQLITE_ROW {
		rv = C.sqlite3_reset(rc.s.s)
		if rv != C.SQLITE_OK {
			return rc.s.c.lastError()
		}
		return nil
	}

	rc.declTypes()

	for i := range dest {
		switch C.sqlite3_column_type(rc.s.s, C.int(i)) {
		case C.SQLITE_INTEGER:
			val := int64(C.sqlite3_column_int64(rc.s.s, C.int(i)))
			switch rc.decltype[i] {
			case columnTimestamp, columnDatetime, columnDate:
				var t time.Time
				// Assume a millisecond unix timestamp if it's 13 digits -- too
				// large to be a reasonable timestamp in seconds.
				if val > 1e12 || val < -1e12 {
					val *= int64(time.Millisecond) // convert ms to nsec
					t = time.Unix(0, val)
				} else {
					t = time.Unix(val, 0)
				}
				t = t.UTC()
				if rc.s.c.tz != nil {
					t = t.In(rc.s.c.tz)
				}
				dest[i] = t
			case "boolean":
				dest[i] = val > 0
			default:
				dest[i] = val
			}
		case C.SQLITE_FLOAT:
			dest[i] = float64(C.sqlite3_column_double(rc.s.s, C.int(i)))
		case C.SQLITE_BLOB:
			p := C.sqlite3_column_blob(rc.s.s, C.int(i))
			if p == nil {
				dest[i] = nil
				continue
			}
			n := int(C.sqlite3_column_bytes(rc.s.s, C.int(i)))
			switch dest[i].(type) {
			default:
				slice := make([]byte, n)
				copy(slice[:], (*[1 << 30]byte)(p)[0:n])
				dest[i] = slice
			}
		case C.SQLITE_NULL:
			dest[i] = nil
		case C.SQLITE_TEXT:
			var err error
			var timeVal time.Time

			n := int(C.sqlite3_column_bytes(rc.s.s, C.int(i)))
			s := C.GoStringN((*C.char)(unsafe.Pointer(C.sqlite3_column_text(rc.s.s, C.int(i)))), C.int(n))

			switch rc.decltype[i] {
			case columnTimestamp, columnDatetime, columnDate:
				var t time.Time
				s = strings.TrimSuffix(s, "Z")
				for _, format := range SQLiteTimestampFormats {
					if timeVal, err = time.ParseInLocation(format, s, time.UTC); err == nil {
						t = timeVal
						break
					}
				}
				if err != nil {
					// The column is a time value, so return the zero time on parse failure.
					t = time.Time{}
				}
				if rc.s.c.tz != nil {
					t = t.In(rc.s.c.tz)
				}
				dest[i] = t
			default:
				dest[i] = []byte(s)
			}

		}
	}
	return nil
}

// ColumnTypeDatabaseTypeName implement RowsColumnTypeDatabaseTypeName.
func (rc *SQLiteRows) ColumnTypeDatabaseTypeName(i int) string {
	return strings.ToUpper(C.GoString(C.sqlite3_column_decltype(rc.s.s, C.int(i))))
}

// ColumnTypeNullable implement RowsColumnTypeNullable.
func (rc *SQLiteRows) ColumnTypeNullable(i int) (nullable, ok bool) {
	return true, true
}

// ColumnTypeScanType implement RowsColumnTypeScanType.
// This can only be successfull retrieved while within Next()
// The underlying SQLite function depends upon sqlite_step to be called.
func (rc *SQLiteRows) ColumnTypeScanType(i int) reflect.Type {
	switch C.sqlite3_column_type(rc.s.s, C.int(i)) {
	case C.SQLITE_INTEGER:
		switch strings.ToLower(C.GoString(C.sqlite3_column_decltype(rc.s.s, C.int(i)))) {
		case "timestamp", "datetime", "date":
			return reflect.TypeOf(time.Time{})
		case "boolean":
			return reflect.TypeOf(false)
		}
		return reflect.TypeOf(int64(0))
	case C.SQLITE_FLOAT:
		return reflect.TypeOf(float64(0))
	case C.SQLITE_BLOB:
		return reflect.SliceOf(reflect.TypeOf(byte(0)))
	case C.SQLITE_NULL:
		return reflect.TypeOf(nil)
	case C.SQLITE_TEXT:
		return reflect.TypeOf("")
	}
	return reflect.SliceOf(reflect.TypeOf(byte(0)))
}
