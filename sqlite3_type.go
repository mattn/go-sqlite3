// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h"
#else
#include <sqlite3.h>
#endif
*/
import "C"
import (
	"database/sql"
	"reflect"
	"time"
)

var (
	type_int      = reflect.TypeOf(int64(0))
	type_float    = reflect.TypeOf(float64(0))
	type_string   = reflect.TypeOf("")
	type_rawbytes = reflect.TypeOf(sql.RawBytes{})
	type_bool     = reflect.TypeOf(true)
	type_time     = reflect.TypeOf(time.Time{})
	type_any      = reflect.TypeOf(new(any)).Elem()
)

// ColumnTypeDatabaseTypeName implement RowsColumnTypeDatabaseTypeName.
func (rc *SQLiteRows) ColumnTypeDatabaseTypeName(i int) string {
	return C.GoString(C.sqlite3_column_decltype(rc.s.s, C.int(i)))
}

/*
func (rc *SQLiteRows) ColumnTypeLength(index int) (length int64, ok bool) {
	return 0, false
}

func (rc *SQLiteRows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return 0, 0, false
}
*/

// ColumnTypeNullable implement RowsColumnTypeNullable.
func (rc *SQLiteRows) ColumnTypeNullable(i int) (nullable, ok bool) {
	return true, true
}

// ColumnTypeScanType implement RowsColumnTypeScanType.
// In SQLite3, this method should be called after Next() has been called, as sqlite3_column_type()
// returns the column type for a specific row. If Next() has not been called, fallback to
// sqlite3_column_decltype()
func (rc *SQLiteRows) ColumnTypeScanType(i int) reflect.Type {
	rc.s.mu.Lock()
	defer rc.s.mu.Unlock()

	if isValidRow := C.sqlite3_stmt_busy(rc.s.s) != 0; !isValidRow {
		return type_any
	}
	if isValidColumn := i >= 0 && i < int(rc.nc); !isValidColumn {
		return type_any
	}

	switch C.sqlite3_column_type(rc.s.s, C.int(i)) {
	case C.SQLITE_INTEGER:
		switch rc.decltype[i] {
		case columnTimestamp, columnDatetime, columnDate:
			return type_time
		case columnBoolean:
			return type_bool
		}
		return type_int
	case C.SQLITE_FLOAT:
		return type_float
	case C.SQLITE_TEXT:
		switch rc.decltype[i] {
		case columnTimestamp, columnDatetime, columnDate:
			return type_time
		}
		return type_string
	case C.SQLITE_BLOB:
		return type_rawbytes
	case C.SQLITE_NULL:
		fallthrough
	default:
		return type_any
	}
}
