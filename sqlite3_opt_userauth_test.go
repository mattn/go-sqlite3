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
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	// Dummy Query to force connection
	if _, err := db.Exec("SELECT 1;"); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}

	// Add normal user to database
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
	db, err = sql.Open("sqlite3", "file:"+tempFilename+"?_auth_user=user&_auth_pass=user")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Add User should now fail because we are not admin
	var rv int
	if err := db.QueryRow("select auth_user_add('user2', 'user2', false);").Scan(&rv); err != nil || rv == 0 {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("Succeeded creating user, while not being admin, this is not supposed to work")
	}

	// Try to create admin user
	// Should also fail because we are not admin
	if err := db.QueryRow("select auth_user_add('admin2', 'admin2', true);").Scan(&rv); err != nil || rv == 0 {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("Succeeded creating admin, while not being admin, this is not supposed to work")
	}
}

func TestAuthorizationFailed(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	// Dummy Query to force connection
	if _, err := db.Exec("SELECT 1;"); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}
	db.Close()

	db, err = sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=invalid")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Dummy Query to issue connection
	if _, err := db.Exec("SELECT 1;"); err != nil && err != ErrUnauthorized {
		t.Fatalf("Failed to connect: %s", err)
	}
}

func TestAuthUserModify(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	var rv int

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	// Dummy Query to force connection
	if _, err := db.Exec("SELECT 1;"); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}

	if err := db.QueryRow("select auth_user_add('user', 'user', false);").Scan(&rv); err != nil || rv != 0 {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("Failed to create normal user")
	}

	if err := db.QueryRow("select auth_user_change('admin', 'nimda', true);").Scan(&rv); err != nil || rv != 0 {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal("Failed to change password")
	}
	db.Close()

	// Re-Connect with new credentials
	db, err = sql.Open("sqlite3", "file:"+tempFilename+"?_auth_user=admin&_auth_pass=nimda")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	if err := db.QueryRow("select count(uname) from sqlite_user where uname = 'admin';").Scan(&rv); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Dummy Query to force connection to test authorization
	if _, err := db.Exec("SELECT 1;"); err != nil && err != ErrUnauthorized {
		t.Fatalf("Failed to connect: %s", err)
	}
}

func TestAuthUserDelete(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	var rv int

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Dummy Query to force connection to test authorization
	if _, err := db.Exec("SELECT 1;"); err != nil {
		t.Fatalf("Failed to connect: %s", err)
	}

	// Add User
	if _, err := db.Exec("select auth_user_add('user', 'user', false);"); err != nil {
		t.Fatal(err)
	}

	// Verify, their should be now 2 users
	var users int
	if err := db.QueryRow("select count(uname) from sqlite_user;").Scan(&users); err != nil {
		t.Fatal(err)
	}
	if users != 2 {
		t.Fatal("Failed to add user")
	}

	// Delete User
	if _, err := db.Exec("select auth_user_delete('user');"); err != nil {
		t.Fatal(err)
	}

	// Verify their should now only be 1 user remaining, the current logged in admin user
	if err := db.QueryRow("select count(uname) from sqlite_user;").Scan(&users); err != nil {
		t.Fatal(err)
	}
	if users != 1 {
		t.Fatal("Failed to delete user")
	}
}
