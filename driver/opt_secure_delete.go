// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build sqlite_secure_delete

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_SECURE_DELETE=1
#cgo LDFLAGS: -lm
*/
import "C"
