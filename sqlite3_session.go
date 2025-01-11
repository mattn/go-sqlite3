//go:build !sqlite_omit_session
// +build !sqlite_omit_session

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_SESSION -DSQLITE_ENABLE_PREUPDATE_HOOK
#cgo LDFLAGS: -lm
*/

/*
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h"
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
*/
import "C"

// CreateSession creates a new session object.
func (c *SQLiteConn) CreateSession() error {
	return nil
}

// AttachSession attaches a session object to a set of tables.
func (c *SQLiteConn) AttachSession() error {
	return nil
}

// DetachSession deletes a session object.
func (c *SQLiteConn) DeleteSession() error {
	return nil
}

// SessionChangeset generates a changeset from a session object.
func (c *SQLiteConn) Changeset() error {
	return nil
}

// ChangesetStart is called to create and initialize an iterator
// to iterate through the contents of a changeset. Initially, the
// iterator points to no element at all
func (c *SQLiteConn) ChangesetStart() error {
	return nil
}

// ChangesetNext moves a Changeset iterator to the next change in the
// changeset.
func (c *SQLiteConn) ChangesetNext() error {
	return nil
}

// ChangesetOp retuns the type of change (INSERT, UPDATE or DELETE)
// that the iterator points to
func (c *SQLiteConn) ChangesetOp() error {
	return nil
}

// ChangesetOld may be used to obtain the old.* values within the change payload.
func (c *SQLiteConn) ChangesetOld() error {
	return nil
}

// ChangesetNew may be used to obtain the new.* values within the change payload.
func (c *SQLiteConn) ChangesetNew() error {
	return nil
}

// ChangesetFinalize is called to delete a changeste iterator.
func (c *SQLiteConn) ChangesetFinalize() error {
	return nil
}

