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
	"strings"
)

const (
	SQLITE_INTEGER = iota
	SQLITE_TEXT
	SQLITE_BLOB
	SQLITE_REAL
	SQLITE_NUMERIC
	SQLITE_TIME
	SQLITE_BOOL
	SQLITE_NULL
)

var (
	TYPE_NULLINT    = reflect.TypeOf(sql.NullInt64{})
	TYPE_NULLFLOAT  = reflect.TypeOf(sql.NullFloat64{})
	TYPE_NULLSTRING = reflect.TypeOf(sql.NullString{})
	TYPE_RAWBYTES   = reflect.TypeOf(sql.RawBytes{})
	TYPE_NULLBOOL   = reflect.TypeOf(sql.NullBool{})
	TYPE_NULLTIME   = reflect.TypeOf(sql.NullTime{})
	TYPE_ANY        = reflect.TypeOf(new(any))
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
	switch C.sqlite3_column_type(rc.s.s, C.int(i)) {
	case C.SQLITE_INTEGER:
		return TYPE_NULLINT
	case C.SQLITE_FLOAT:
		return TYPE_NULLFLOAT
	case C.SQLITE_TEXT:
		return TYPE_NULLSTRING
	case C.SQLITE_BLOB:
		return TYPE_RAWBYTES
		// This case can signal that the value is NULL or that Next() has not been called yet.
		// Skip it and return the fallback behaviour as a best effort. This is safe as all types
		// returned are Nullable or any, which is the expected value for SQLite3.
		//case C.SQLITE_NULL:
		//	return TYPE_ANY
	}

	// Fallback to schema declared to remain retro-compatible
	return scanType(C.GoString(C.sqlite3_column_decltype(rc.s.s, C.int(i))))
}

func scanType(cdt string) reflect.Type {
	t := strings.ToUpper(cdt)
	i := databaseTypeConvSqlite(t)
	switch i {
	case SQLITE_INTEGER:
		return TYPE_NULLINT
	case SQLITE_TEXT:
		return TYPE_NULLSTRING
	case SQLITE_BLOB:
		return TYPE_RAWBYTES
	case SQLITE_REAL:
		return TYPE_NULLFLOAT
	case SQLITE_NUMERIC:
		return TYPE_NULLFLOAT
	case SQLITE_BOOL:
		return TYPE_NULLBOOL
	case SQLITE_TIME:
		return TYPE_NULLTIME
	}
	return TYPE_ANY
}

func databaseTypeConvSqlite(t string) int {
	if strings.Contains(t, "INT") {
		return SQLITE_INTEGER
	}
	if t == "CLOB" || t == "TEXT" ||
		strings.Contains(t, "CHAR") {
		return SQLITE_TEXT
	}
	if t == "BLOB" {
		return SQLITE_BLOB
	}
	if t == "REAL" || t == "FLOAT" ||
		strings.Contains(t, "DOUBLE") {
		return SQLITE_REAL
	}
	if t == "DATE" || t == "DATETIME" ||
		t == "TIMESTAMP" {
		return SQLITE_TIME
	}
	if t == "NUMERIC" ||
		strings.Contains(t, "DECIMAL") {
		return SQLITE_NUMERIC
	}
	if t == "BOOLEAN" {
		return SQLITE_BOOL
	}

	return SQLITE_NULL
}
