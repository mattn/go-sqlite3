// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !sqlite_userauth

package sqlite3

import (
	"C"
)

// Authenticate will perform an authentication of the provided username
// and password against the database.
//
// If a database contains the SQLITE_USER table, then the
// call to Authenticate must be invoked with an
// appropriate username and password prior to enable read and write
//access to the database.
//
// Return SQLITE_OK on success or SQLITE_ERROR if the username/password
// combination is incorrect or unknown.
//
// If the SQLITE_USER table is not present in the database file, then
// this interface is a harmless no-op returnning SQLITE_OK.
func (c *SQLiteConn) Authenticate(username, password string) error {
	// NOOP
	return nil
}

// AuthUserAdd can be used (by an admin user only)
// to create a new user.  When called on a no-authentication-required
// database, this routine converts the database into an authentication-
// required database, automatically makes the added user an
// administrator, and logs in the current connection as that user.
// The AuthUserAdd only works for the "main" database, not
// for any ATTACH-ed databases. Any call to AuthUserAdd by a
// non-admin user results in an error.
func (c *SQLiteConn) AuthUserAdd(username, password string, admin bool) error {
	// NOOP
	return nil
}

// AuthUserChange can be used to change a users
// login credentials or admin privilege.  Any user can change their own
// login credentials.  Only an admin user can change another users login
// credentials or admin privilege setting.  No user may change their own
// admin privilege setting.
func (c *SQLiteConn) AuthUserChange(username, password string, admin bool) error {
	// NOOP
	return nil
}

// AuthUserDelete can be used (by an admin user only)
// to delete a user.  The currently logged-in user cannot be deleted,
// which guarantees that there is always an admin user and hence that
// the database cannot be converted into a no-authentication-required
// database.
func (c *SQLiteConn) AuthUserDelete(username string) error {
	// NOOP
	return nil
}

// EOF
