//go:build sqlite_session
// +build sqlite_session

package sqlite3

// The Session Extension
// https://sqlite.org/sessionintro.html

/*
#cgo CFLAGS: -DSQLITE_ENABLE_SESSION
#cgo CFLAGS: -DSQLITE_ENABLE_PREUPDATE_HOOK
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>

static int xConflict(void *pCtx, int eConflict, sqlite3_changeset_iter *pIter){
  int ret = (int)pCtx;
  return ret;
}

int apply_changeset(
  sqlite3 *db,
  int bIgnoreConflicts,
  int nChangeset,
  void *pChangeset
){
  return sqlite3changeset_apply(
      db,
      nChangeset, pChangeset,
      0, xConflict,
      (void*)bIgnoreConflicts
  );
}
*/
import "C"

import (
	"unsafe"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// https://sqlite.org/session/constlist.html
const (
	SQLITE_CHANGESETAPPLY_FKNOACTION  = C.SQLITE_CHANGESETAPPLY_FKNOACTION
	SQLITE_CHANGESETAPPLY_IGNORENOOP  = C.SQLITE_CHANGESETAPPLY_IGNORENOOP
	SQLITE_CHANGESETAPPLY_INVERT      = C.SQLITE_CHANGESETAPPLY_INVERT
	SQLITE_CHANGESETAPPLY_NOSAVEPOINT = C.SQLITE_CHANGESETAPPLY_NOSAVEPOINT
	SQLITE_CHANGESETSTART_INVERT      = C.SQLITE_CHANGESETSTART_INVERT
	SQLITE_CHANGESET_ABORT            = C.SQLITE_CHANGESET_ABORT
	SQLITE_CHANGESET_CONFLICT         = C.SQLITE_CHANGESET_CONFLICT
	SQLITE_CHANGESET_CONSTRAINT       = C.SQLITE_CHANGESET_CONSTRAINT
	SQLITE_CHANGESET_DATA             = C.SQLITE_CHANGESET_DATA
	SQLITE_CHANGESET_FOREIGN_KEY      = C.SQLITE_CHANGESET_FOREIGN_KEY
	SQLITE_CHANGESET_NOTFOUND         = C.SQLITE_CHANGESET_NOTFOUND
	SQLITE_CHANGESET_OMIT             = C.SQLITE_CHANGESET_OMIT
	SQLITE_CHANGESET_REPLACE          = C.SQLITE_CHANGESET_REPLACE
	SQLITE_SESSION_CONFIG_STRMSIZE    = C.SQLITE_SESSION_CONFIG_STRMSIZE
	SQLITE_SESSION_OBJCONFIG_ROWID    = C.SQLITE_SESSION_OBJCONFIG_ROWID
	SQLITE_SESSION_OBJCONFIG_SIZE     = C.SQLITE_SESSION_OBJCONFIG_SIZE
)

// https://sqlite.org/session/session.html
type SQLiteSession struct {
	db  *C.sqlite3
	ptr *C.sqlite3_session
}

// https://sqlite.org/session/changegroup.html
type SQLiteChangegroup struct {
	db  *C.sqlite3
	ptr *C.sqlite3_changegroup
}

// https://sqlite.org/session/rebaser.html
type SQLiteRebaser struct {
	db  *C.sqlite3
	ptr *C.sqlite3_rebaser
}

// https://sqlite.org/session/changeset_iter.html
type SQLiteChangesetIterator struct {
	db  *C.sqlite3
	ptr *C.sqlite3_changeset_iter
}

// int sqlite3changegroup_add(sqlite3_changegroup*, int nData, void *pData);
func (cg *SQLiteChangegroup) Add(nData int, pData unsafe.Pointer) error {
	rc := C.sqlite3changegroup_add(cg.ptr, C.int(nData), pData)
	if rc != C.SQLITE_OK {
		return lastError(cg.db)
	}
	return nil
}

// int sqlite3changegroup_add_change(
//
//	sqlite3_changegroup*,
//	sqlite3_changeset_iter*
//
// );
func (cg *SQLiteChangegroup) AddChange(iter *SQLiteChangesetIterator) error {
	rc := C.sqlite3changegroup_add_change(cg.ptr, iter.ptr)
	if rc != C.SQLITE_OK {
		return lastError(cg.db)
	}
	return nil
}

// void sqlite3changegroup_delete(sqlite3_changegroup*);
func (cg *SQLiteChangegroup) Delete() {
	C.sqlite3changegroup_delete(cg.ptr)
}

// int sqlite3changegroup_new(sqlite3_changegroup **pp);
func (c *SQLiteConn) ChangegroupNew() (*SQLiteChangegroup, error) {
	var cg *C.sqlite3_changegroup
	rc := C.sqlite3changegroup_new(&cg)
	if rc != C.SQLITE_OK {
		return nil, lastError(c.db)
	}
	cgS := &SQLiteChangegroup{db: c.db, ptr: cg}
	return cgS, nil
}

// int sqlite3changegroup_output(
//
//	sqlite3_changegroup*,
//	int *pnData,                    /* OUT: Size of output buffer in bytes */
//	void **ppData                   /* OUT: Pointer to output buffer */
//
// );
func (cg *SQLiteChangegroup) Output() ([]byte, error) {
	var pnData C.int
	var ppData unsafe.Pointer
	rc := C.sqlite3changegroup_output(cg.ptr, &pnData, &ppData)
	if rc != C.SQLITE_OK {
		return nil, lastError(cg.db)
	}
	defer C.free(unsafe.Pointer(ppData))
	bytes := C.GoBytes(ppData, pnData)
	if len(bytes) == 0 {
		return nil, nil
	}
	// Copy the bytes to a new slice to avoid memory issues
	// when the C memory is freed.
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// int sqlite3changegroup_schema(sqlite3_changegroup*, sqlite3*, const char *zDb);
func (cg *SQLiteChangegroup) Schema(zDb string) error {
	czDb := C.CString(zDb)
	defer C.free(unsafe.Pointer(czDb))
	rc := C.sqlite3changegroup_schema(cg.ptr, cg.db, czDb)
	if rc != C.SQLITE_OK {
		return lastError(cg.db)
	}
	return nil
}

// int sqlite3changeset_apply(
//
//	sqlite3 *db,                    /* Apply change to "main" db of this handle */
//	int nChangeset,                 /* Size of changeset in bytes */
//	void *pChangeset,               /* Changeset blob */
//	int(*xFilter)(
//	  void *pCtx,                   /* Copy of sixth arg to _apply() */
//	  const char *zTab              /* Table name */
//	),
//	int(*xConflict)(
//	  void *pCtx,                   /* Copy of sixth arg to _apply() */
//	  int eConflict,                /* DATA, MISSING, CONFLICT, CONSTRAINT */
//	  sqlite3_changeset_iter *p     /* Handle describing change and conflict */
//	),
//	void *pCtx                      /* First argument passed to xConflict */
//
// );
func (c *SQLiteConn) ChangesetApply(data []byte, xFilter func(zTab string) int, xConflict func(eConflict int, p *SQLiteChangesetIterator) int) int {
	db := c.db

	var xConflictFunc unsafe.Pointer
	var xFilterFunc unsafe.Pointer

	if xConflict != nil {
		handle := func(pCtx unsafe.Pointer, eConflict C.int, p *C.sqlite3_changeset_iter) C.int {
			iter := SQLiteChangesetIterator{
				db:  db,
				ptr: p,
			}
			return C.int(xConflict(int(eConflict), &iter))
		}
		handlePtr := unsafe.Pointer(&handle)
		xConflictFunc = handlePtr
	} else {
		// Provide a default conflict handler that always returns SQLITE_CHANGESET_OMIT (0)
		defaultConflictHandler := func(pCtx unsafe.Pointer, eConflict C.int, p *C.sqlite3_changeset_iter) C.int {
			return C.SQLITE_CHANGESET_OMIT
		}
		handlePtr := unsafe.Pointer(&defaultConflictHandler)
		xConflictFunc = handlePtr
	}

	if xFilter != nil {
		handle := func(pCtx unsafe.Pointer, zTab *C.char) C.int {
			goTab := C.GoString(zTab)
			return C.int(xFilter(goTab))
		}
		handlePtr := unsafe.Pointer(&handle)
		xFilterFunc = handlePtr
	} else {
		xFilterFunc = nil
	}

	cBytes := unsafe.Pointer(&data[0])
	cLen := C.int(len(data))

	rc := C.sqlite3changeset_apply(
		db,
		cLen,
		cBytes,
		(*[0]byte)(xFilterFunc),
		(*[0]byte)(xConflictFunc),
		nil,
	)

	return int(rc)
}

// int sqlite3changeset_apply_v2(
//
//	sqlite3 *db,                    /* Apply change to "main" db of this handle */
//	int nChangeset,                 /* Size of changeset in bytes */
//	void *pChangeset,               /* Changeset blob */
//	int(*xFilter)(
//	  void *pCtx,                   /* Copy of sixth arg to _apply() */
//	  const char *zTab              /* Table name */
//	),
//	int(*xConflict)(
//	  void *pCtx,                   /* Copy of sixth arg to _apply() */
//	  int eConflict,                /* DATA, MISSING, CONFLICT, CONSTRAINT */
//	  sqlite3_changeset_iter *p     /* Handle describing change and conflict */
//	),
//	void *pCtx,                     /* First argument passed to xConflict */
//	void **ppRebase, int *pnRebase, /* OUT: Rebase data */
//	int flags                       /* SESSION_CHANGESETAPPLY_* flags */
//
// );
func (c *SQLiteConn) ChangesetApplyV2(data []byte, xFilter func(zTab string) int, xConflict func(eConflict int, p *SQLiteChangesetIterator) int, flags int) ([]byte, error) {
	db := c.db

	var xConflictFunc unsafe.Pointer
	var xFilterFunc unsafe.Pointer

	if xConflict != nil {
		handle := func(pCtx unsafe.Pointer, eConflict C.int, p *C.sqlite3_changeset_iter) C.int {
			iter := SQLiteChangesetIterator{
				db:  db,
				ptr: p,
			}
			return C.int(xConflict(int(eConflict), &iter))
		}
		handlePtr := unsafe.Pointer(&handle)
		xConflictFunc = handlePtr
	} else {
		// Provide a default conflict handler that always returns SQLITE_CHANGESET_OMIT (0)
		defaultConflictHandler := func(pCtx unsafe.Pointer, eConflict C.int, p *C.sqlite3_changeset_iter) C.int {
			return C.SQLITE_CHANGESET_OMIT
		}
		handlePtr := unsafe.Pointer(&defaultConflictHandler)
		xConflictFunc = handlePtr
	}

	if xFilter != nil {
		handle := func(pCtx unsafe.Pointer, zTab *C.char) C.int {
			goTab := C.GoString(zTab)
			return C.int(xFilter(goTab))
		}
		handlePtr := unsafe.Pointer(&handle)
		xFilterFunc = handlePtr
	} else {
		xFilterFunc = nil
	}

	cBytes := unsafe.Pointer(&data[0])
	cLen := C.int(len(data))

	var ppRebase unsafe.Pointer
	var pnRebase C.int

	rc := C.sqlite3changeset_apply_v2(
		db,
		cLen,
		cBytes,
		(*[0]byte)(xFilterFunc),
		(*[0]byte)(xConflictFunc),
		nil,
		&ppRebase,
		&pnRebase,
		C.int(flags),
	)
	if rc != C.SQLITE_OK {
		return nil, lastError(db)
	}
	if ppRebase == nil {
		return nil, nil
	}
	defer C.free(unsafe.Pointer(ppRebase))
	bytes := C.GoBytes(ppRebase, pnRebase)
	if len(bytes) == 0 {
		return nil, nil
	}
	// Copy the bytes to a new slice to avoid memory issues
	// when the C memory is freed.
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// int sqlite3changeset_concat(
//
//	int nA,                         /* Number of bytes in buffer pA */
//	void *pA,                       /* Pointer to buffer containing changeset A */
//	int nB,                         /* Number of bytes in buffer pB */
//	void *pB,                       /* Pointer to buffer containing changeset B */
//	int *pnOut,                     /* OUT: Number of bytes in output changeset */
//	void **ppOut                    /* OUT: Buffer containing output changeset */
//
// );
func (c *SQLiteConn) ChangesetConcat(dataA []byte, dataB []byte) ([]byte, error) {
	db := c.db
	var pnOut C.int
	var ppOut unsafe.Pointer
	defer C.free(ppOut)
	cDataA := C.CBytes(dataA)
	defer C.free(cDataA)
	cDataB := C.CBytes(dataB)
	defer C.free(cDataB)
	rc := C.sqlite3changeset_concat(
		C.int(len(dataA)),
		cDataA,
		C.int(len(dataB)),
		cDataB,
		&pnOut,
		&ppOut,
	)
	if rc != C.SQLITE_OK {
		return nil, lastError(db)
	}
	if int(pnOut) == 0 {
		return nil, nil
	}
	bytes := C.GoBytes(ppOut, pnOut)
	if len(bytes) == 0 {
		return nil, nil
	}
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// int sqlite3changeset_conflict(
//
//	sqlite3_changeset_iter *pIter,  /* Changeset iterator */
//	int iVal,                       /* Column number */
//	sqlite3_value **ppValue         /* OUT: Value from conflicting row */
//
// );
func (itr *SQLiteChangesetIterator) ChangesetConflict(iVal int) (*C.sqlite3_value, error) {
	var ppValue *C.sqlite3_value
	rc := C.sqlite3changeset_conflict(itr.ptr, C.int(iVal), &ppValue)
	if rc != C.SQLITE_OK {
		return nil, lastError(itr.db)
	}
	if ppValue == nil {
		return nil, nil
	}
	// Copy the value to a new pointer to avoid memory issues
	// when the C memory is freed.
	value := (*C.sqlite3_value)(unsafe.Pointer(ppValue))
	return value, nil
}

// int sqlite3changeset_finalize(sqlite3_changeset_iter *pIter);
func (itr *SQLiteChangesetIterator) Finalize() {
	C.sqlite3changeset_finalize(itr.ptr)
}

// int sqlite3changeset_fk_conflicts(
//
//	sqlite3_changeset_iter *pIter,  /* Changeset iterator */
//	int *pnOut                      /* OUT: Number of FK violations */
//
// );
func (itr *SQLiteChangesetIterator) FkConflicts() (int, error) {
	var pnOut C.int
	rc := C.sqlite3changeset_fk_conflicts(itr.ptr, &pnOut)
	if rc != C.SQLITE_OK {
		return 0, lastError(itr.db)
	}
	return int(pnOut), nil
}

// int sqlite3changeset_invert(
//
//	int nIn, const void *pIn,       /* Input changeset */
//	int *pnOut, void **ppOut        /* OUT: Inverse of input */
//
// );
func (c *SQLiteConn) ChangesetInvert(data []byte) ([]byte, error) {
	db := c.db
	var pnOut C.int
	var ppOut unsafe.Pointer
	defer C.free(ppOut)
	cData := C.CBytes(data)
	defer C.free(cData)
	rc := C.sqlite3changeset_invert(
		C.int(len(data)),
		cData,
		&pnOut,
		&ppOut,
	)
	if rc != C.SQLITE_OK {
		return nil, lastError(db)
	}
	if int(pnOut) == 0 {
		return nil, nil
	}
	bytes := C.GoBytes(ppOut, pnOut)
	if len(bytes) == 0 {
		return nil, nil
	}
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// int sqlite3changeset_new(
//
//	sqlite3_changeset_iter *pIter,  /* Changeset iterator */
//	int iVal,                       /* Column number */
//	sqlite3_value **ppValue         /* OUT: New value (or NULL pointer) */
//
// );
func (itr *SQLiteChangesetIterator) ChangesetNew(iVal int) (*C.sqlite3_value, error) {
	var ppValue *C.sqlite3_value
	rc := C.sqlite3changeset_new(itr.ptr, C.int(iVal), &ppValue)
	if rc != C.SQLITE_OK {
		return nil, lastError(itr.db)
	}
	if ppValue == nil {
		return nil, nil
	}
	// Copy the value to a new pointer to avoid memory issues
	// when the C memory is freed.
	value := (*C.sqlite3_value)(unsafe.Pointer(ppValue))
	return value, nil
}

// int sqlite3changeset_next(sqlite3_changeset_iter *pIter);
func (itr *SQLiteChangesetIterator) Next() error {
	rc := C.sqlite3changeset_next(itr.ptr)
	if rc != C.SQLITE_OK {
		return lastError(itr.db)
	}
	return nil
}

// int sqlite3changeset_old(
//
//	sqlite3_changeset_iter *pIter,  /* Changeset iterator */
//	int iVal,                       /* Column number */
//	sqlite3_value **ppValue         /* OUT: Old value (or NULL pointer) */
//
// );
func (itr *SQLiteChangesetIterator) ChangesetOld(iVal int) (*C.sqlite3_value, error) {
	var ppValue *C.sqlite3_value
	rc := C.sqlite3changeset_old(itr.ptr, C.int(iVal), &ppValue)
	if rc != C.SQLITE_OK {
		return nil, lastError(itr.db)
	}
	if ppValue == nil {
		return nil, nil
	}
	// Copy the value to a new pointer to avoid memory issues
	// when the C memory is freed.
	value := (*C.sqlite3_value)(unsafe.Pointer(ppValue))
	return value, nil
}

// int sqlite3changeset_op(
//
//	sqlite3_changeset_iter *pIter,  /* Iterator object */
//	const char **pzTab,             /* OUT: Pointer to table name */
//	int *pnCol,                     /* OUT: Number of columns in table */
//	int *pOp,                       /* OUT: SQLITE_INSERT, DELETE or UPDATE */
//	int *pbIndirect                 /* OUT: True for an 'indirect' change */
//
// );

type SQLiteChangesetOp struct {
	TableName string
	Column    int
	Op        int
	Indirect  bool
}

func (itr *SQLiteChangesetIterator) ChangesetOp() (*SQLiteChangesetOp, error) {
	var pzTab *C.char
	var pnCol C.int
	var pOp C.int
	var pbIndirect C.int
	rc := C.sqlite3changeset_op(itr.ptr, &pzTab, &pnCol, &pOp, &pbIndirect)
	if rc != C.SQLITE_OK {
		return nil, lastError(itr.db)
	}
	tab := C.GoString(pzTab)
	op := int(pOp)
	col := int(pnCol)
	indirect := pbIndirect == 1
	result := &SQLiteChangesetOp{
		TableName: tab,
		Column:    col,
		Op:        op,
		Indirect:  indirect,
	}
	return result, nil
}

// int sqlite3changeset_pk(
//
//	sqlite3_changeset_iter *pIter,  /* Iterator object */
//	unsigned char **pabPK,          /* OUT: Array of boolean - true for PK cols */
//	int *pnCol                      /* OUT: Number of entries in output array */
//
// );
func (itr *SQLiteChangesetIterator) ChangesetPk() ([]bool, error) {
	var pabPK *C.uchar
	var pnCol C.int
	rc := C.sqlite3changeset_pk(itr.ptr, &pabPK, &pnCol)
	if rc != C.SQLITE_OK {
		return nil, lastError(itr.db)
	}
	if pabPK == nil {
		return nil, nil
	}
	pk := make([]bool, pnCol)
	for i := 0; i < int(pnCol); i++ {
		val := *(*C.uchar)(unsafe.Pointer(uintptr(unsafe.Pointer(pabPK)) + uintptr(i)))
		if val == 1 {
			pk[i] = true
		} else {
			pk[i] = false
		}
	}
	return pk, nil
}

// int sqlite3changeset_start(
//
//	sqlite3_changeset_iter **pp,    /* OUT: New changeset iterator handle */
//	int nChangeset,                 /* Size of changeset blob in bytes */
//	void *pChangeset                /* Pointer to blob containing changeset */
//
// );
func (c *SQLiteConn) ChangesetStart(data []byte) (*SQLiteChangesetIterator, error) {
	db := c.db
	var pp *C.sqlite3_changeset_iter
	cData := C.CBytes(data)
	defer C.free(cData)
	rc := C.sqlite3changeset_start(&pp, C.int(len(data)), cData)
	if rc != C.SQLITE_OK {
		return nil, lastError(db)
	}
	itr := &SQLiteChangesetIterator{db: db, ptr: pp}
	return itr, nil
}

// int sqlite3changeset_start_v2(
//
//	sqlite3_changeset_iter **pp,    /* OUT: New changeset iterator handle */
//	int nChangeset,                 /* Size of changeset blob in bytes */
//	void *pChangeset,               /* Pointer to blob containing changeset */
//	int flags                       /* SESSION_CHANGESETSTART_* flags */
//
// );
func (c *SQLiteConn) ChangesetStartV2(data []byte, flags int) (*SQLiteChangesetIterator, error) {
	db := c.db
	var pp *C.sqlite3_changeset_iter
	cData := C.CBytes(data)
	defer C.free(cData)
	rc := C.sqlite3changeset_start_v2(&pp, C.int(len(data)), cData, C.int(flags))
	if rc != C.SQLITE_OK {
		return nil, lastError(db)
	}
	itr := &SQLiteChangesetIterator{db: db, ptr: pp}
	return itr, nil
}

// // int sqlite3changeset_upgrade(
// //
// //	sqlite3 *db,
// //	const char *zDb,
// //	int nIn, const void *pIn,       /* Input changeset */
// //	int *pnOut, void **ppOut        /* OUT: Inverse of input */
// //
// // );
// func (c *SQLiteConn) ChangesetUpgrade(db string, data []byte) ([]byte, error) {
// 	cDb := C.CString(db)
// 	defer C.free(unsafe.Pointer(cDb))
// 	var pnOut C.int
// 	var ppOut unsafe.Pointer
// 	defer C.free(ppOut)
// 	cData := C.CBytes(data)
// 	defer C.free(cData)
// 	rc := C.sqlite3changeset_upgrade(
// 		c.db,
// 		cDb,
// 		C.int(len(data)),
// 		cData,
// 		&pnOut,
// 		&ppOut,
// 	)
// 	if rc != C.SQLITE_OK {
// 		return nil, c.lastError()
// 	}
// 	if int(pnOut) == 0 {
// 		return nil, nil
// 	}
// 	bytes := C.GoBytes(ppOut, pnOut)
// 	if len(bytes) == 0 {
// 		return nil, nil
// 	}
// 	result := make([]byte, len(bytes))
// 	copy(result, bytes)
// 	return result, nil
// }

// int sqlite3rebaser_configure(
//
//	sqlite3_rebaser*,
//	int nRebase, const void *pRebase
//
// );
func (r *SQLiteRebaser) Configure(data []byte) error {
	cData := C.CBytes(data)
	defer C.free(cData)
	rc := C.sqlite3rebaser_configure(r.ptr, C.int(len(data)), cData)
	if rc != C.SQLITE_OK {
		return lastError(r.db)
	}
	return nil
}

// int sqlite3rebaser_create(sqlite3_rebaser **ppNew);
func (c *SQLiteConn) RebaserCreate() (*SQLiteRebaser, error) {
	db := c.db
	var rebaser *C.sqlite3_rebaser
	rc := C.sqlite3rebaser_create(&rebaser)
	if rc != C.SQLITE_OK {
		return nil, lastError(db)
	}
	r := &SQLiteRebaser{db: db, ptr: rebaser}
	return r, nil
}

// void sqlite3rebaser_delete(sqlite3_rebaser *p);
func (r *SQLiteRebaser) Delete() {
	C.sqlite3rebaser_delete(r.ptr)
}

// int sqlite3rebaser_rebase(
//
//	sqlite3_rebaser*,
//	int nIn, const void *pIn,
//	int *pnOut, void **ppOut
//
// );
func (r *SQLiteRebaser) Rebase(data []byte) ([]byte, error) {
	var pnOut C.int
	var ppOut unsafe.Pointer
	defer C.free(ppOut)
	cData := C.CBytes(data)
	defer C.free(cData)
	rc := C.sqlite3rebaser_rebase(
		r.ptr,
		C.int(len(data)),
		cData,
		&pnOut,
		&ppOut,
	)
	if rc != C.SQLITE_OK {
		return nil, lastError(r.db)
	}
	if int(pnOut) == 0 {
		return nil, nil
	}
	bytes := C.GoBytes(ppOut, pnOut)
	if len(bytes) == 0 {
		return nil, nil
	}
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// int sqlite3session_create(
//
//	sqlite3 *db,                    /* Database handle */
//	const char *zDb,                /* Name of db (e.g. "main") */
//	sqlite3_session **ppSession     /* OUT: New session object */
//
// );
func (c *SQLiteConn) SessionCreate(name string) (*SQLiteSession, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var session *C.sqlite3_session
	rc := C.sqlite3session_create(c.db, cname, &session)
	if rc != C.SQLITE_OK {
		return nil, lastError(c.db)
	}
	s := &SQLiteSession{db: c.db, ptr: session}
	return s, nil
}

// void sqlite3session_delete(sqlite3_session *pSession);
func (s *SQLiteSession) Delete() {
	C.sqlite3session_delete(s.ptr)
}

// int sqlite3session_enable(sqlite3_session *pSession, int bEnable);
func (s *SQLiteSession) Enable(enable bool) bool {
	rc := C.sqlite3session_enable(s.ptr, C.int(boolToInt(enable)))
	intValue := int(rc)
	return intValue == 1
}

// int sqlite3session_indirect(sqlite3_session *pSession, int bIndirect);
func (s *SQLiteSession) SetIndirect(enable bool) int {
	rc := C.sqlite3session_indirect(s.ptr, C.int(boolToInt(enable)))
	intValue := int(rc)
	return intValue
}

// int sqlite3session_indirect(sqlite3_session *pSession, int bIndirect);
func (s *SQLiteSession) IsIndirect() bool {
	rc := C.sqlite3session_indirect(s.ptr, C.int(-1))
	return rc == 1
}

// int sqlite3session_isempty(sqlite3_session *pSession);
func (s *SQLiteSession) IsEmpty() bool {
	rc := C.sqlite3session_isempty(s.ptr)
	return rc == 1
}

// int sqlite3session_attach(
//
//	sqlite3_session *pSession,      /* Session object */
//	const char *zTab                /* Table name */
//
// );
func (s *SQLiteSession) Attach(tableName string) error {
	cname := C.CString(tableName)
	defer C.free(unsafe.Pointer(cname))
	if tableName == "" {
		cname = nil
	}
	rc := C.sqlite3session_attach(s.ptr, cname)
	if rc != C.SQLITE_OK {
		return lastError(s.db)
	}
	return nil
}

// sqlite3_int64 sqlite3session_changeset_size(sqlite3_session *pSession);
func (s *SQLiteSession) ChangesetSize() int64 {
	rc := C.sqlite3session_changeset_size(s.ptr)
	return int64(rc)
}

// sqlite3_int64 sqlite3session_memory_used(sqlite3_session *pSession);
func (s *SQLiteSession) ChangesetMemoryUsed() int64 {
	rc := C.sqlite3session_memory_used(s.ptr)
	return int64(rc)
}

// int sqlite3session_diff(
//
//	sqlite3_session *pSession,
//	const char *zFromDb,
//	const char *zTbl,
//	char **pzErrMsg
//
// );
func (s *SQLiteSession) Diff(fromDb string, tableName string) ([]byte, error) {
	cfromDb := C.CString(fromDb)
	ctableName := C.CString(tableName)
	defer C.free(unsafe.Pointer(cfromDb))
	defer C.free(unsafe.Pointer(ctableName))
	var pzErrMsg *C.char
	rc := C.sqlite3session_diff(s.ptr, cfromDb, ctableName, &pzErrMsg)
	if rc != C.SQLITE_OK {
		err := lastError(s.db)
		if pzErrMsg != nil {
			C.free(unsafe.Pointer(pzErrMsg))
		}
		return nil, err
	}
	return nil, nil
}

// int sqlite3session_config(int op, void *pArg);
func SessionConfig(op int, pArg unsafe.Pointer) error {
	rc := C.sqlite3session_config(C.int(op), pArg)
	if rc != C.SQLITE_OK {
		return lastError(nil)
	}
	return nil
}

// int sqlite3session_changeset(
//
//	sqlite3_session *pSession,      /* Session object */
//	int *pnChangeset,               /* OUT: Size of buffer at *ppChangeset */
//	void **ppChangeset              /* OUT: Buffer containing changeset */
//
// );
func (s *SQLiteSession) Changeset() ([]byte, error) {
	var pnChangeset C.int
	var ppChangeset unsafe.Pointer
	defer C.free(unsafe.Pointer(ppChangeset))
	rc := C.sqlite3session_changeset(s.ptr, &pnChangeset, &ppChangeset)
	if rc != C.SQLITE_OK {
		return nil, lastError(s.db)
	}
	bytes := C.GoBytes(ppChangeset, pnChangeset)
	if len(bytes) == 0 {
		return nil, nil
	}
	// Copy the bytes to a new slice to avoid memory issues
	// when the C memory is freed.
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// int sqlite3session_patchset(
//
//	sqlite3_session *pSession,      /* Session object */
//	int *pnPatchset,                /* OUT: Size of buffer at *ppPatchset */
//	void **ppPatchset               /* OUT: Buffer containing patchset */
//
// );
func (s *SQLiteSession) Patchset() ([]byte, error) {
	var pnPatchset C.int
	var ppPatchset unsafe.Pointer
	defer C.free(unsafe.Pointer(ppPatchset))
	rc := C.sqlite3session_patchset(s.ptr, &pnPatchset, &ppPatchset)
	if rc != C.SQLITE_OK {
		return nil, lastError(s.db)
	}
	bytes := C.GoBytes(ppPatchset, pnPatchset)
	if len(bytes) == 0 {
		return nil, nil
	}
	// Copy the bytes to a new slice to avoid memory issues
	// when the C memory is freed.
	result := make([]byte, len(bytes))
	copy(result, bytes)
	return result, nil
}

// void sqlite3session_table_filter(
//
//	sqlite3_session *pSession,      /* Session object */
//	int(*xFilter)(
//	  void *pCtx,                   /* Copy of third arg to _filter_table() */
//	  const char *zTab              /* Table name */
//	),
//	void *pCtx                      /* First argument passed to xFilter */
//
// );
func (s *SQLiteSession) TableFilter(xFilter func(zTab string) int) error {
	var xFilterFunc *[0]byte
	ctx := unsafe.Pointer(s.db)
	defer C.free(ctx)
	if xFilter != nil {
		handle := func(zTab string) int {
			return xFilter(zTab)
		}
		handlePtr := unsafe.Pointer(&handle)
		xFilterFunc = (*[0]byte)(handlePtr)
	} else {
		xFilterFunc = nil
	}
	C.sqlite3session_table_filter(s.ptr, xFilterFunc, ctx)
	return nil
}
