// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

func TestForeignKeys(t *testing.T) {
	cases := map[string]bool{
		"?foreign_keys=1": true,
		"?foreign_keys=0": false,
	}
	for option, want := range cases {
		fname := TempFilename(t)
		uri := "file:" + fname + option
		db, err := sql.Open("sqlite3", uri)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uri, err)
			continue
		}
		var enabled bool
		err = db.QueryRow("PRAGMA foreign_keys;").Scan(&enabled)
		db.Close()
		os.Remove(fname)
		if err != nil {
			t.Errorf("query foreign_keys for %s: %v", uri, err)
			continue
		}
		if enabled != want {
			t.Errorf("\"PRAGMA foreign_keys;\" for %q = %t; want %t", uri, enabled, want)
			continue
		}
	}
}
