// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build solaris

package sqlite3

/*
#cgo CFLAGS: -D__EXTENSIONS__=1
#cgo LDFLAGS: -lc
*/
import "C"
