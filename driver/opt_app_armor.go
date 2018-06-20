// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build !windows
// +build sqlite_app_armor

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_API_ARMOR
#cgo LDFLAGS: -lm
*/
import "C"
