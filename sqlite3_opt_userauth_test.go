// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_userauth

package sqlite3

import (
	"database/sql"
	"fmt"
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

	var exists bool
	err = db.QueryRow("select count(type) from sqlite_master WHERE type='table' and name='sqlite_user';").Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("failed to enable User Authentication")
	}
}

func TestAuthorization(t *testing.T) {
	tempFilename := TempFilename(t)
	fmt.Println(tempFilename)
	//defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	if _, err := db.Exec("select auth_user_add('user', 'user', false);"); err != nil {
		t.Fatal(err)
	}

	var uname string
	if err := db.QueryRow("select uname from sqlite_user where uname = 'user';").Scan(&uname); err != nil {
		t.Fatal(err)
	}

	if uname != "user" {
		t.Fatal("Failed to create normal user")
	}
	db.Close()

	// Re-Open Database as User
	// Add User should now fail because we are not admin
	db, err = sql.Open("sqlite3", "file:"+tempFilename+"?_auth_user=user&_auth_pass=user")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Try to create normal user
	var rv string
	if err := db.QueryRow("select auth_user_add('user2', 'user2', false);").Scan(&rv); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("RV: %v\n", rv)
	// if rv != SQLITE_AUTH {
	// 	t.Fatal("Succeeded creating user while not admin")
	// }

	// // Try to create admin user
	// if _, err := db.Exec("select auth_user_add('admin2', 'admin2', true);"); err != nil {
	// 	t.Fatal(err)
	// }
}
