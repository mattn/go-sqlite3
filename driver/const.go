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
	// SQLITE_DELETE authorizer action code
	SQLITE_DELETE = C.SQLITE_DELETE

	// SQLITE_INSERT authorizer action code
	SQLITE_INSERT = C.SQLITE_INSERT

	// SQLITE_UPDATE authorizer action code
	SQLITE_UPDATE = C.SQLITE_UPDATE
)

const (
	columnDate      string = "date"
	columnDatetime  string = "datetime"
	columnTimestamp string = "timestamp"
)
