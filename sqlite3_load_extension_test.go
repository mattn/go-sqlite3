// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !sqlite_omit_load_extension

package sqlite3

import (
	"database/sql"
	"testing"
)

func TestExtensionsError(t *testing.T) {
	sql.Register("sqlite3_TestExtensionsError",
		&SQLiteDriver{
			Extensions: []string{
				"foobar",
			},
		},
	)

	db, err := sql.Open("sqlite3_TestExtensionsError", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		t.Fatal("expected error loading non-existent extension")
	}

	if err.Error() == "not an error" {
		t.Fatal("expected error from sqlite3_enable_load_extension to be returned")
	}
}

func TestLoadExtensionError(t *testing.T) {
	sql.Register("sqlite3_TestLoadExtensionError",
		&SQLiteDriver{
			ConnectHook: func(c *SQLiteConn) error {
				return c.LoadExtension("foobar", "")
			},
		},
	)

	db, err := sql.Open("sqlite3_TestLoadExtensionError", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		t.Fatal("expected error loading non-existent extension")
	}

	if err.Error() == "not an error" {
		t.Fatal("expected error from sqlite3_enable_load_extension to be returned")
	}
}
