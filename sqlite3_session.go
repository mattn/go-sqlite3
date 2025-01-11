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
#include <stdint.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type Session struct {
	session *C.sqlite3_session
}

type Changeset struct {
	changeset *C.sqlite3_changeset_iter
}

// CreateSession creates a new session object.
func (c *SQLiteConn) CreateSession(dbName string) (*Session, error) {
	cDbName := C.CString(dbName)
	defer C.free(unsafe.Pointer(cDbName))

	var session *C.sqlite3_session
	rc := C.sqlite3session_create(c.db, cDbName, &session)
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3session_create: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	return &Session{session: session}, nil
}

// AttachSession attaches a session object to a table or all tables.
func (s *Session) AttachSession(tableName string) error {
	var cTableName *C.char
	if tableName != "" {
		cTableName = C.CString(tableName)
		defer C.free(unsafe.Pointer(cTableName))
	}

	rc := C.sqlite3session_attach(s.session, cTableName)
	if rc != C.SQLITE_OK {
		return fmt.Errorf("sqlite3session_attach: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	return nil
}

// Delete deletes a session object.
func (s *Session) DeleteSession() error {
	if s.session != nil {
		// Call sqlite3session_delete to free the session object
		C.sqlite3session_delete(s.session)
		s.session = nil // Set session to nil to avoid double deletion
	}
	return nil
}

type ChangesetIter struct {
	iter *C.sqlite3_changeset_iter
}

// Changeset generates a changeset from a session object.
func (s *Session) Changeset() ([]byte, error) {
	var nChangeset C.int
	var pChangeset unsafe.Pointer

	// Call sqlite3session_changeset
	rc := C.sqlite3session_changeset(s.session, &nChangeset, &pChangeset)
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3session_changeset: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	defer C.sqlite3_free(pChangeset) // Free the changeset buffer after use

	// Convert the C buffer to a Go byte slice
	changeset := C.GoBytes(pChangeset, nChangeset)
	return changeset, nil
}

// ChangesetStart creates and initializes a changeset iterator.
func ChangesetStart(changeset []byte) (*ChangesetIter, error) {
	var iter *C.sqlite3_changeset_iter

	// Call sqlite3changeset_start
	rc := C.sqlite3changeset_start(&iter, C.int(len(changeset)), unsafe.Pointer(&changeset[0]))
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3changeset_start: %s", C.GoString(C.sqlite3_errstr(rc)))
	}

	return &ChangesetIter{iter: iter}, nil
}

// ChangesetNext moves the changeset iterator to the next change.
func (ci *ChangesetIter) ChangesetNext() (bool, error) {
	rc := C.sqlite3changeset_next(ci.iter)
	if rc == C.SQLITE_DONE {
		return false, nil // No more changes
	}
	if rc != C.SQLITE_OK {
		return false, fmt.Errorf("sqlite3changeset_next: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	return true, nil
}

// ChangesetOp returns the type of change (INSERT, UPDATE, or DELETE) that the iterator points to.
func (ci *ChangesetIter) ChangesetOp() (string, int, int, bool, error) {
	var tableName *C.char
	var nCol C.int
	var op C.int
	var indirect C.int

	rc := C.sqlite3changeset_op(ci.iter, &tableName, &nCol, &op, &indirect)
	if rc != C.SQLITE_OK {
		return "", 0, 0, false, fmt.Errorf("sqlite3changeset_op: %s", C.GoString(C.sqlite3_errstr(rc)))
	}

	return C.GoString(tableName), int(nCol), int(op), indirect != 0, nil
}

// ChangesetOld retrieves the old value for the specified column in the change payload.
func (ci *ChangesetIter) ChangesetOld(column int) (*C.sqlite3_value, error) {
	var value *C.sqlite3_value

	rc := C.sqlite3changeset_old(ci.iter, C.int(column), &value)
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3changeset_old: %s", C.GoString(C.sqlite3_errstr(rc)))
	}

	return value, nil
}

// ChangesetNew retrieves the new value for the specified column in the change payload.
func (ci *ChangesetIter) ChangesetNew(column int) (*C.sqlite3_value, error) {
	var value *C.sqlite3_value

	rc := C.sqlite3changeset_new(ci.iter, C.int(column), &value)
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3changeset_new: %s", C.GoString(C.sqlite3_errstr(rc)))
	}

	return value, nil
}

// ChangesetFinalize deletes a changeset iterator.
func (ci *ChangesetIter) ChangesetFinalize() error {
	if ci.iter != nil {
		rc := C.sqlite3changeset_finalize(ci.iter)
		ci.iter = nil // Prevent double finalization
		if rc != C.SQLITE_OK {
			return fmt.Errorf("sqlite3changeset_finalize: %s", C.GoString(C.sqlite3_errstr(rc)))
		}
	}
	return nil
}
