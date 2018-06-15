// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build libsqlite3

package sqlite3

/*
#cgo CFLAGS: -DUSE_LIBSQLITE3
#cgo linux LDFLAGS: -lsqlite3
#cgo darwin LDFLAGS: -L/usr/local/opt/sqlite/lib -lsqlite3
#cgo openbsd LDFLAGS: -lsqlite3
#cgo solaris LDFLAGS: -lsqlite3
*/
import "C"
