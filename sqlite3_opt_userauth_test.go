// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_userauth

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

func TestAuthCreateDatabase(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Ping database
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	var exists bool
	err = db.QueryRow("select count(type) from sqlite_master WHERE type='table' and name='sqlite_user';").Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("failed to enable User Authentication")
	}
}
