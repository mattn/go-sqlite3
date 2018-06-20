// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build sqlite_icu

package sqlite3

/*
#cgo LDFLAGS: -licuuc -licui18n
#cgo CFLAGS: -DSQLITE_ENABLE_ICU
#cgo darwin CFLAGS: -I/usr/local/opt/icu4c/include
#cgo darwin LDFLAGS: -L/usr/local/opt/icu4c/lib
#cgo openbsd LDFLAGS: -lsqlite3
*/
import "C"
