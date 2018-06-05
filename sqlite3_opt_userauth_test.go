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

	. "github.com/smartystreets/goconvey/convey"
)

var (
	conn        *SQLiteConn
	connect     func(t *testing.T, f string, username, password string) (file string, db *sql.DB, c *SQLiteConn, err error)
	authEnabled func(db *sql.DB) (exists bool, err error)
	addUser     func(db *sql.DB, username, password string, admin int) (rv int, err error)
	userExists  func(db *sql.DB, username string) (rv int, err error)
	isAdmin     func(db *sql.DB, username string) (rv bool, err error)
	modifyUser  func(db *sql.DB, username, password string, admin int) (rv int, err error)
	deleteUser  func(db *sql.DB, username string) (rv int, err error)
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

		db, err = sql.Open("sqlite3_with_conn", "file:"+file+fmt.Sprintf("?_auth&_auth_user=%s&_auth_pass=%s", username, password))
		if err != nil {
			defer os.Remove(file)
			return file, nil, nil, err
		}

		// Dummy query to force connection and database creation
		// Will return ErrUnauthorized (SQLITE_AUTH) if user authentication fails
		if _, err = db.Exec("SELECT 1;"); err != nil {
			defer os.Remove(file)
			defer db.Close()
			return file, nil, nil, err
		}
		c = conn

		return
	}

	authEnabled = func(db *sql.DB) (exists bool, err error) {
		err = db.QueryRow("select count(type) from sqlite_master WHERE type='table' and name='sqlite_user';").Scan(&exists)
		return
	}

	addUser = func(db *sql.DB, username, password string, admin int) (rv int, err error) {
		err = db.QueryRow("select auth_user_add(?, ?, ?);", username, password, admin).Scan(&rv)
		return
	}

	userExists = func(db *sql.DB, username string) (rv int, err error) {
		err = db.QueryRow("select count(uname) from sqlite_user where uname=?", username).Scan(&rv)
		return
	}

	isAdmin = func(db *sql.DB, username string) (rv bool, err error) {
		err = db.QueryRow("select isAdmin from sqlite_user where uname=?", username).Scan(&rv)
		return
	}

	modifyUser = func(db *sql.DB, username, password string, admin int) (rv int, err error) {
		err = db.QueryRow("select auth_user_change(?, ?, ?);", username, password, admin).Scan(&rv)
		return
	}

	deleteUser = func(db *sql.DB, username string) (rv int, err error) {
		err = db.QueryRow("select auth_user_delete(?);", username).Scan(&rv)
		return
	}
}

func TestUserAuthentication(t *testing.T) {
	Convey("Create Database", t, func() {
		f, db, c, err := connect(t, "", "admin", "admin")
		So(db, ShouldNotBeNil)
		So(c, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db.Close()
		defer os.Remove(f)

		b, err := authEnabled(db)
		So(b, ShouldEqual, true)
		So(err, ShouldBeNil)

		e, err := userExists(db, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
	})

	Convey("Authorization Success", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connect(t, f1, "admin", "admin")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("Authorization Success (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)
		defer db1.Close()

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Test lower level authentication
		err = c1.Authenticate("admin", "admin")
		So(err, ShouldBeNil)
	})

	Convey("Authorization Failed", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Perform Invalid Authentication when we connect
		// to a database
		f2, db2, c2, err := connect(t, f1, "admin", "invalid")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldBeNil)
		So(c2, ShouldBeNil)
		So(err, ShouldEqual, ErrUnauthorized)
	})

	Convey("Authorization Failed (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)
		defer db1.Close()

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Test lower level authentication
		// We require a successful *SQLiteConn to test this.
		err = c1.Authenticate("admin", "invalid")
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrUnauthorized)
	})
}

func TestUserAuthenticationAddUser(t *testing.T) {
	Convey("Add Admin User", t, func() {
		f, db, c, err := connect(t, "", "admin", "admin")
		So(db, ShouldNotBeNil)
		So(c, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f)
		defer db.Close()

		e, err := userExists(db, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Admin User
		rv, err := addUser(db, "admin2", "admin2", 1)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		e, err = userExists(db, "admin2")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db, "admin2")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
	})

	Convey("Add Admin User (*SQLiteConn)", t, func() {
		f, db, c, err := connect(t, "", "admin", "admin")
		So(db, ShouldNotBeNil)
		So(c, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f)
		defer db.Close()

		e, err := userExists(db, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Test lower level AuthUserAdd
		err = c.AuthUserAdd("admin2", "admin2", true)
		So(err, ShouldBeNil)

		e, err = userExists(db, "admin2")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db, "admin2")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
	})

	Convey("Add Normal User", t, func() {
		f, db, c, err := connect(t, "", "admin", "admin")
		So(db, ShouldNotBeNil)
		So(c, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f)
		defer db.Close()

		e, err := userExists(db, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		e, err = userExists(db, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
	})

	Convey("Add Normal User (*SQLiteConn)", t, func() {
		f, db, c, err := connect(t, "", "admin", "admin")
		So(db, ShouldNotBeNil)
		So(c, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f)
		defer db.Close()

		e, err := userExists(db, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Test lower level AuthUserAdd
		err = c.AuthUserAdd("user", "user", false)
		So(err, ShouldBeNil)

		e, err = userExists(db, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
	})

	Convey("Add Admin User Insufficient Privileges", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		// Add Admin User
		// Because 'user' is not admin
		// Adding an admin user should now fail
		// because we have insufficient privileges
		rv, err = addUser(db2, "admin2", "admin2", 1)
		So(rv, ShouldEqual, SQLITE_AUTH)
		So(err, ShouldBeNil)
	})

	Convey("Add Admin User Insufficient Privileges (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		// Add Admin User
		// Because 'user' is not admin
		// Adding an admin user should now fail
		// because we have insufficient privileges
		err = c2.AuthUserAdd("admin2", "admin2", true)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrAdminRequired)
	})

	Convey("Add Normal User Insufficient Privileges", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		// Add Normal User
		// Because 'user' is not admin
		// Adding an normal user should now fail
		// because we have insufficient privileges
		rv, err = addUser(db2, "user2", "user2", 0)
		So(rv, ShouldEqual, SQLITE_AUTH)
		So(err, ShouldBeNil)
	})

	Convey("Add Normal User Insufficient Privileges (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		// Add Normal User
		// Because 'user' is not admin
		// Adding an normal user should now fail
		// because we have insufficient privileges
		// Test lower level AuthUserAdd
		err = c2.AuthUserAdd("user2", "user2", false)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrAdminRequired)
	})
}

func TestUserAuthenticationModifyUser(t *testing.T) {
	Convey("Modify Current Connection Password", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Modify Password
		rv, err := modifyUser(db1, "admin", "admin2", 1)
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, 0)
		db1.Close()

		// Reconnect with new password
		f2, db2, c2, err := connect(t, f1, "admin", "admin2")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("Modify Current Connection Password (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)
		defer db1.Close()

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Modify password through (*SQLiteConn)
		err = c1.AuthUserChange("admin", "admin2", true)
		So(err, ShouldBeNil)

		// Reconnect with new password
		f2, db2, c2, err := connect(t, f1, "admin", "admin2")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("Modify Current Connection Admin Flag", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)
		defer db1.Close()

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Modify Administrator Flag
		// Because we are current logged in as 'admin'
		// Changing our own admin flag should fail.
		rv, err := modifyUser(db1, "admin", "admin", 0)
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, SQLITE_AUTH)
	})

	Convey("Modify Current Connection Admin Flag (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)
		defer db1.Close()

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Modify admin flag through (*SQLiteConn)
		// Because we are current logged in as 'admin'
		// Changing our own admin flag should fail.
		err = c1.AuthUserChange("admin", "admin", false)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrAdminRequired)
	})

	Convey("Modify Other User Password", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify User
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Modify Password for user
		rv, err = modifyUser(db1, "user", "user2", 0)
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, 0)
		db1.Close()

		// Reconnect as normal user with new password
		f2, db2, c2, err := connect(t, f1, "user", "user2")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("Modify Other User Password (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify User
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Modify user password through (*SQLiteConn)
		// Because we are still logged in as admin
		// this should succeed.
		err = c1.AuthUserChange("admin", "admin", false)
		So(err, ShouldNotBeNil)
	})

	Convey("Modify Other User Admin Flag", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify User
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Modify Password for user
		// Because we are logged in as admin
		// This call should succeed.
		rv, err = modifyUser(db1, "user", "user", 1)
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, 0)
		db1.Close()

		// Reconnect as normal user with new password
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("Modify Other User Admin Flag (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify User
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Modify user password through (*SQLiteConn)
		// Because we are still logged in as admin
		// this should succeed.
		err = c1.AuthUserChange("user", "user", true)
		So(err, ShouldBeNil)
	})

	Convey("Modify Other User Password as Non-Admin", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Add Normal User
		rv, err = addUser(db1, "user2", "user2", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify 'user'
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Verify 'user2'
		e, err = userExists(db1, "user2")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user2")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		// Modify password for user as normal user
		// Because 'user' is not admin
		// Modifying password as a normal user should now fail
		// because we have insufficient privileges
		rv, err = modifyUser(db2, "user2", "invalid", 0)
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, SQLITE_AUTH)
	})

	Convey("Modify Other User Password as Non-Admin", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Add Normal User
		rv, err = addUser(db1, "user2", "user2", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify 'user'
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Verify 'user2'
		e, err = userExists(db1, "user2")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user2")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		// Modify user password through (*SQLiteConn)
		// for 'user2'
		// Because 'user' is not admin
		// Modifying password as a normal user should now fail
		// because we have insufficient privileges
		err = c2.AuthUserChange("user2", "invalid", false)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrAdminRequired)
	})
}

func TestUserAuthenticationDeleteUser(t *testing.T) {
	Convey("Delete User as Admin", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify 'user'
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		rv, err = deleteUser(db1, "user")
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, 0)

		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 0)
	})

	Convey("Delete User as Admin (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify 'user'
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		err = c1.AuthUserDelete("user")
		So(err, ShouldBeNil)

		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 0)
	})

	Convey("Delete User as Non-Admin", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Add Normal User
		rv, err = addUser(db1, "user2", "user2", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify 'user'
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Verify 'user2'
		e, err = userExists(db1, "user2")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user2")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		rv, err = deleteUser(db2, "user2")
		So(err, ShouldBeNil)
		So(rv, ShouldEqual, SQLITE_AUTH)
	})

	Convey("Delete User as Non-Admin (*SQLiteConn)", t, func() {
		f1, db1, c1, err := connect(t, "", "admin", "admin")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)

		// Add Normal User
		rv, err := addUser(db1, "user", "user", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Add Normal User
		rv, err = addUser(db1, "user2", "user2", 0)
		So(rv, ShouldEqual, 0) // 0 == C.SQLITE_OK
		So(err, ShouldBeNil)

		// Verify 'user'
		e, err = userExists(db1, "user")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)

		// Verify 'user2'
		e, err = userExists(db1, "user2")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err = isAdmin(db1, "user2")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, false)
		db1.Close()

		// Reconnect as normal user
		f2, db2, c2, err := connect(t, f1, "user", "user")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()

		err = c2.AuthUserDelete("user2")
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrAdminRequired)
	})
}

func TestUserAuthenticationEncoder(t *testing.T) {
	connectWithCrypt := func(t *testing.T, f string, username, password string, crypt string, salt string) (file string, db *sql.DB, c *SQLiteConn, err error) {
		conn = nil // Clear connection
		file = f   // Copy provided file (f) => file
		if file == "" {
			// Create dummy file
			file = TempFilename(t)
		}

		db, err = sql.Open("sqlite3_with_conn", "file:"+file+fmt.Sprintf("?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=%s&_auth_salt=%s", username, password, crypt, salt))
		if err != nil {
			defer os.Remove(file)
			return file, nil, nil, err
		}

		// Dummy query to force connection and database creation
		// Will return ErrUnauthorized (SQLITE_AUTH) if user authentication fails
		if _, err = db.Exec("SELECT 1;"); err != nil {
			defer os.Remove(file)
			defer db.Close()
			return file, nil, nil, err
		}
		c = conn

		return
	}

	Convey("SHA1 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "sha1", "")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "sha1", "")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SSHA1 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "ssha1", "salted")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "ssha1", "salted")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SHA256 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "sha256", "")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "sha256", "")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SSHA256 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "ssha256", "salted")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "ssha256", "salted")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SHA384 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "sha384", "")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "sha384", "")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SSHA384 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "ssha384", "salted")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "ssha384", "salted")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SHA512 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "sha512", "")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "sha512", "")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})

	Convey("SSHA512 Encoder", t, func() {
		f1, db1, c1, err := connectWithCrypt(t, "", "admin", "admin", "ssha512", "salted")
		So(f1, ShouldNotBeBlank)
		So(db1, ShouldNotBeNil)
		So(c1, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer os.Remove(f1)

		e, err := userExists(db1, "admin")
		So(err, ShouldBeNil)
		So(e, ShouldEqual, 1)

		a, err := isAdmin(db1, "admin")
		So(err, ShouldBeNil)
		So(a, ShouldEqual, true)
		db1.Close()

		// Preform authentication
		f2, db2, c2, err := connectWithCrypt(t, f1, "admin", "admin", "ssha512", "salted")
		So(f2, ShouldNotBeBlank)
		So(f1, ShouldEqual, f2)
		So(db2, ShouldNotBeNil)
		So(c2, ShouldNotBeNil)
		So(err, ShouldBeNil)
		defer db2.Close()
	})
}
