// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build !sqlite_userauth

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

// This file is included for code coverage

func TestUserAuthOmit(t *testing.T) {
	var conn *SQLiteConn

	sql.Register("sqlite3_with_conn",
		&SQLiteDriver{
			ConnectHook: func(c *SQLiteConn) error {
				conn = c
				return nil
			},
		})

	file := TempFilename(t)
	defer os.Remove(file)

	db, err := sql.Open("sqlite3_with_conn", "file:"+file+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("UserAuthOmit Failure")
	}
	defer db.Close()

	// Dummy query to force connection and database creation
	// Will return ErrUnauthorized (SQLITE_AUTH) if user authentication fails
	if _, err := db.Exec("SELECT 1;"); err != nil {
		t.Fatal("UserAuthOmit Failure")
	}

	if conn.Authenticate("", ""); err != nil {
		t.Fatal("UserAuthOmit Failure")
	}
	if ok := conn.authenticate("", ""); ok != 0 {
		t.Fatal("UserAuthOmit Failure")
	}

	if conn.AuthUserAdd("", "", true); err != nil {
		t.Fatal("UserAuthOmit Failure")
	}
	if ok := conn.authUserAdd("", "", 1); ok != 0 {
		t.Fatal("UserAuthOmit Failure")
	}

	if conn.AuthUserChange("", "", true); err != nil {
		t.Fatal("UserAuthOmit Failure")
	}
	if ok := conn.authUserChange("", "", 1); ok != 0 {
		t.Fatal("UserAuthOmit Failure")
	}

	if conn.AuthUserDelete(""); err != nil {
		t.Fatal("UserAuthOmit Failure")
	}
	if ok := conn.authUserDelete(""); ok != 0 {
		t.Fatal("UserAuthOmit Failure")
	}
	if enabled := conn.AuthEnabled(); enabled {
		t.Fatal("UserAuthOmit Failure")
	}
}
