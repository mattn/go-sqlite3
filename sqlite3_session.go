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
		C.sqlite3session_delete(s.session)
		s.session = nil
	}
	return nil
}

// Changeset represents a changeset object.
type Changeset struct {
	b []byte
}

// NewChangeset returns a changeset from a session object.
func NewChangeset(s *Session) (*Changeset, error) {
	var nChangeset C.int
	var pChangeset unsafe.Pointer

	rc := C.sqlite3session_changeset(s.session, &nChangeset, &pChangeset)
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3session_changeset: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	defer C.sqlite3_free(pChangeset)

	// Copy the changeset buffer to a Go byte slice, because cgo
	// does not support Go slices with C memory.
	changeset := C.GoBytes(pChangeset, nChangeset)
	return &Changeset{b: changeset}, nil
}

// ChangesetIterator represents a changeset iterator object.
type ChangesetIterator struct {
	iter *C.sqlite3_changeset_iter
}

// NewChangesetIterator creates a new changeset iterator object.
func NewChangesetIterator(cs *Changeset) (*ChangesetIterator, error) {
	var iter *C.sqlite3_changeset_iter
	ptr := unsafe.Pointer(nil)
	if len(cs.b) > 0 {
		ptr = unsafe.Pointer(&cs.b[0])
	}
	rc := C.sqlite3changeset_start(&iter, C.int(len(cs.b)), ptr)
	if rc != C.SQLITE_OK {
		return nil, fmt.Errorf("sqlite3changeset_start: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	return &ChangesetIterator{iter: iter}, nil
}

// Next moves the changeset iterator to the next change.
func (ci *ChangesetIterator) Next() (bool, error) {
	rc := C.sqlite3changeset_next(ci.iter)
	if rc == C.SQLITE_DONE {
		return false, nil // No more changes
	}
	if rc != C.SQLITE_ROW {
		return false, fmt.Errorf("sqlite3changeset_next: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	return true, nil
}

// Op returns the current Operation from a Changeset Iterator
func (ci *ChangesetIterator) Op() (tblName string, numCol int, oper int, indirect bool, err error) {
	var tableName *C.char
	var nCol C.int
	var op C.int
	var ind C.int

	rc := C.sqlite3changeset_op(ci.iter, &tableName, &nCol, &op, &ind)
	if rc != C.SQLITE_OK {
		return "", 0, 0, false, fmt.Errorf("sqlite3changeset_op: %s", C.GoString(C.sqlite3_errstr(rc)))
	}
	return C.GoString(tableName), int(nCol), int(op), ind != 0, nil
}

// Old retrieves the old value for the specified column in the change payload.
func (ci *ChangesetIterator) Old(dest []any) error {
	return ci.row(dest, true)
}

// New retrieves the new value for the specified column in the change payload.
func (ci *ChangesetIterator) New(dest []any) error {
	return ci.row(dest, false)
}

// Finalize deletes a changeset iterator.
func (ci *ChangesetIterator) Finalize() error {
	if ci.iter != nil {
		rc := C.sqlite3changeset_finalize(ci.iter)
		ci.iter = nil
		if rc != C.SQLITE_OK {
			return fmt.Errorf("sqlite3changeset_finalize: %s", C.GoString(C.sqlite3_errstr(rc)))
		}
	}
	return nil
}

// New retrieves the new value for the specified column in the change payload.
func (ci *ChangesetIterator) row(dest []any, old bool) error {
	var val *C.sqlite3_value
	var rc C.int
	for i := 0; i < len(dest); i++ {
		fn := ""
		if old {
			fn = "old"
			rc = C.sqlite3changeset_old(ci.iter, C.int(i), &val)
		} else {
			fn = "new"
			rc = C.sqlite3changeset_new(ci.iter, C.int(i), &val)
		}
		if rc != C.SQLITE_OK {
			return fmt.Errorf("sqlite3changeset_%s: %s", fn, C.GoString(C.sqlite3_errstr(rc)))
		}

		switch C.sqlite3_value_type(val) {
		case C.SQLITE_INTEGER:
			dest[i] = int64(C.sqlite3_value_int64(val))
		case C.SQLITE_FLOAT:
			dest[i] = float64(C.sqlite3_value_double(val))
		case C.SQLITE_BLOB:
			len := C.sqlite3_value_bytes(val)
			blobptr := C.sqlite3_value_blob(val)
			dest[i] = C.GoBytes(blobptr, len)
		case C.SQLITE_TEXT:
			cstrptr := unsafe.Pointer(C.sqlite3_value_text(val))
			dest[i] = C.GoString((*C.char)(cstrptr))
		case C.SQLITE_NULL:
			dest[i] = nil
		}
	}
	return nil
}
