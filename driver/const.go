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
*/
import "C"

const (
	// SQLiteDelete authorizer action code
	SQLiteDelete = C.SQLITE_DELETE

	// SQLiteInsert authorizer action code
	SQLiteInsert = C.SQLITE_INSERT

	// SQLiteUpdate authorizer action code
	SQLiteUpdate = C.SQLITE_UPDATE
)

const (
	// SQLITE_DELETE authorizer action code
	//
	// Deprecated: Use SQLiteDelete instead.
	SQLITE_DELETE = SQLiteDelete

	// SQLITE_INSERT authorizer action code
	//
	// Deprecated: Use SQLiteInsert instead.
	SQLITE_INSERT = SQLiteInsert

	// SQLITE_UPDATE authorizer action code
	//
	// Deprecated: Use SQLiteUpdate instead.
	SQLITE_UPDATE = SQLiteUpdate
)

const (
	columnDate      string = "date"
	columnDatetime  string = "datetime"
	columnTimestamp string = "timestamp"
)
