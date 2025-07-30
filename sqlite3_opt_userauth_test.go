// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_userauth
// +build sqlite_userauth

package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
)

var (
	conn    *SQLiteConn
	connect func(t *testing.T, f string, username, password string) (file string, db *sql.DB, c *SQLiteConn, err error)
)

func init() {
	// Create database connection
	sql.Register("sqlite3_with_conn",
		&SQLiteDriver{
			ConnectHook: func(c *SQLiteConn) error {
				conn = c
				return nil
			},
		})

	connect = func(t *testing.T, f string, username, password string) (file string, db *sql.DB, c *SQLiteConn, err error) {
		conn = nil // Clear connection
		file = f   // Copy provided file (f) => file
		if file == "" {
			// Create dummy file
			file = TempFilename(t)
		}

		params := "?_auth"
		if len(username) > 0 {
			params = fmt.Sprintf("%s&_auth_user=%s", params, username)
		}
		if len(password) > 0 {
			params = fmt.Sprintf("%s&_auth_pass=%s", params, password)
		}
		db, err = sql.Open("sqlite3_with_conn", "file:"+file+params)
		if err != nil {
			defer os.Remove(file)
			return file, nil, nil, err
		}

		// Dummy query to force connection and database creation
		// Will return errUserAuthNoLongerSupported if user authentication fails
		if _, err = db.Exec("SELECT 1;"); err != nil {
			defer os.Remove(file)
			defer db.Close()
			return file, nil, nil, err
		}
		c = conn

		return
	}
}

func TestUserAuth(t *testing.T) {
	_, _, _, err := connect(t, "", "admin", "admin")
	if err == nil {
		t.Fatalf("UserAuth not enabled: %s", err)
	}
	if !errors.Is(err, errUserAuthNoLongerSupported) {
		t.Fatal(err)
	}
}
