// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_userauth

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_USER_AUTHENTICATION
#cgo LDFLAGS: -lm
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>

static int
_sqlite3_user_authenticate(sqlite3* db, const char* zUsername, const char* aPW, int nPW)
{
  return sqlite3_user_authenticate(db, zUsername, aPW, nPW);
}

static int
_sqlite3_user_add(sqlite3* db, const char* zUsername, const char* aPW, int nPW, int isAdmin)
{
  return sqlite3_user_add(db, zUsername, aPW, nPW, isAdmin);
}

static int
_sqlite3_user_change(sqlite3* db, const char* zUsername, const char* aPW, int nPW, int isAdmin)
{
  return sqlite3_user_change(db, zUsername, aPW, nPW, isAdmin);
}

static int
_sqlite3_user_delete(sqlite3* db, const char* zUsername)
{
  return sqlite3_user_delete(db, zUsername);
}
*/
import "C"

const (
	SQLITE_AUTH = C.SQLITE_AUTH
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
	rv := C._sqlite3_user_authenticate(c.db, C.CString(username), C.CString(password), C.int(len(password)))
	if rv != C.SQLITE_OK {
		return c.lastError()
	}

	return nil
}

// AuthUserAdd can be used (by an admin user only)
// to create a new user. When called on a no-authentication-required
// database, this routine converts the database into an authentication-
// required database, automatically makes the added user an
// administrator, and logs in the current connection as that user.
// The AuthUserAdd only works for the "main" database, not
// for any ATTACH-ed databases. Any call to AuthUserAdd by a
// non-admin user results in an error.
func (c *SQLiteConn) AuthUserAdd(username, password string, admin bool) error {
	isAdmin := 0
	if admin {
		isAdmin = 1
	}

	rv := C._sqlite3_user_add(c.db, C.CString(username), C.CString(password), C.int(len(password)), C.int(isAdmin))
	if rv != C.SQLITE_OK {
		return c.lastError()
	}

	return nil
}

// AuthUserChange can be used to change a users
// login credentials or admin privilege.  Any user can change their own
// login credentials. Only an admin user can change another users login
// credentials or admin privilege setting. No user may change their own
// admin privilege setting.
func (c *SQLiteConn) AuthUserChange(username, password string, admin bool) error {
	isAdmin := 0
	if admin {
		isAdmin = 1
	}

	rv := C._sqlite3_user_change(c.db, C.CString(username), C.CString(password), C.int(len(password)), C.int(isAdmin))
	if rv != C.SQLITE_OK {
		return c.lastError()
	}

	return nil
}

// AuthUserDelete can be used (by an admin user only)
// to delete a user. The currently logged-in user cannot be deleted,
// which guarantees that there is always an admin user and hence that
// the database cannot be converted into a no-authentication-required
// database.
func (c *SQLiteConn) AuthUserDelete(username string) error {
	rv := C._sqlite3_user_delete(c.db, C.CString(username))
	if rv != C.SQLITE_OK {
		return c.lastError()
	}

	return nil
}

// EOF
