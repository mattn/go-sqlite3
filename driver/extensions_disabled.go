// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_omit_load_extension sqlite_disable_extensions

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_OMIT_LOAD_EXTENSION
*/
import "C"
import (
	"errors"
)

func (c *SQLiteConn) loadExtensions(extensions []string) error {
	return errors.New("extensions have been disabled for static builds")
}

func (c *SQLiteConn) LoadExtension(lib string, entry string) error {
	return errors.New("extensions have been disabled for static builds")
}
