// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

/*
#cgo CFLAGS: -std=gnu99
#cgo CFLAGS: -DSQLITE_ENABLE_RTREE -DSQLITE_THREADSAFE
#cgo CFLAGS: -DSQLITE_ENABLE_FTS3 -DSQLITE_ENABLE_FTS3_PARENTHESIS -DSQLITE_ENABLE_FTS4_UNICODE61
#cgo CFLAGS: -DSQLITE_TRACE_SIZE_LIMIT=15
#cgo CFLAGS: -DSQLITE_ENABLE_COLUMN_METADATA=1
#cgo CFLAGS: -Wno-deprecated-declarations

#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
#include <stdint.h>

int goSqlite3CreateModule(sqlite3 *db, const char *zName, uintptr_t pClientData);

static inline char *my_mprintf(char *zFormat, char *arg) {
	return sqlite3_mprintf(zFormat, arg);
}
*/
import "C"

import (
	"math"
	"reflect"
	"unsafe"
)

type sqliteModule struct {
	c      *SQLiteConn
	name   string
	module Module
}

type sqliteVTab struct {
	module *sqliteModule
	vTab   VTab
}

type sqliteVTabCursor struct {
	vTab       *sqliteVTab
	vTabCursor VTabCursor
}

// Op is type of operations.
type Op uint8

// Op mean identity of operations.
const (
	OpEQ         Op = 2
	OpGT            = 4
	OpLE            = 8
	OpLT            = 16
	OpGE            = 32
	OpMATCH         = 64
	OpLIKE          = 65 /* 3.10.0 and later only */
	OpGLOB          = 66 /* 3.10.0 and later only */
	OpREGEXP        = 67 /* 3.10.0 and later only */
	OpScanUnique    = 1  /* Scan visits at most 1 row */
)

// InfoConstraint give information of constraint.
type InfoConstraint struct {
	Column int
	Op     Op
	Usable bool
}

// InfoOrderBy give information of order-by.
type InfoOrderBy struct {
	Column int
	Desc   bool
}

func constraints(info *C.sqlite3_index_info) []InfoConstraint {
	l := info.nConstraint
	slice := (*[1 << 30]C.struct_sqlite3_index_constraint)(unsafe.Pointer(info.aConstraint))[:l:l]

	cst := make([]InfoConstraint, 0, l)
	for _, c := range slice {
		var usable bool
		if c.usable > 0 {
			usable = true
		}
		cst = append(cst, InfoConstraint{
			Column: int(c.iColumn),
			Op:     Op(c.op),
			Usable: usable,
		})
	}
	return cst
}

func orderBys(info *C.sqlite3_index_info) []InfoOrderBy {
	l := info.nOrderBy
	slice := (*[1 << 30]C.struct_sqlite3_index_orderby)(unsafe.Pointer(info.aOrderBy))[:l:l]

	ob := make([]InfoOrderBy, 0, l)
	for _, c := range slice {
		var desc bool
		if c.desc > 0 {
			desc = true
		}
		ob = append(ob, InfoOrderBy{
			Column: int(c.iColumn),
			Desc:   desc,
		})
	}
	return ob
}

// IndexResult is a Go struct represetnation of what eventually ends up in the
// output fields for `sqlite3_index_info`
// See: https://www.sqlite.org/c3ref/index_info.html
type IndexResult struct {
	Used           []bool // aConstraintUsage
	IdxNum         int
	IdxStr         string
	AlreadyOrdered bool // orderByConsumed
	EstimatedCost  float64
	EstimatedRows  float64
}

// mPrintf is a utility wrapper around sqlite3_mprintf
func mPrintf(format, arg string) *C.char {
	cf := C.CString(format)
	defer C.free(unsafe.Pointer(cf))
	ca := C.CString(arg)
	defer C.free(unsafe.Pointer(ca))
	return C.my_mprintf(cf, ca)
}

//export goMInit
func goMInit(db, pClientData unsafe.Pointer, argc int, argv **C.char, pzErr **C.char, isCreate int) C.uintptr_t {
	m := lookupHandle(uintptr(pClientData)).(*sqliteModule)
	if m.c.db != (*C.sqlite3)(db) {
		*pzErr = mPrintf("%s", "Inconsistent db handles")
		return 0
	}
	args := make([]string, argc)
	var A []*C.char
	slice := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(argv)), Len: argc, Cap: argc}
	a := reflect.NewAt(reflect.TypeOf(A), unsafe.Pointer(&slice)).Elem().Interface()
	for i, s := range a.([]*C.char) {
		args[i] = C.GoString(s)
	}
	var vTab VTab
	var err error
	if isCreate == 1 {
		vTab, err = m.module.Create(m.c, args)
	} else {
		vTab, err = m.module.Connect(m.c, args)
	}

	if err != nil {
		*pzErr = mPrintf("%s", err.Error())
		return 0
	}
	vt := sqliteVTab{m, vTab}
	*pzErr = nil
	return C.uintptr_t(newHandle(m.c, &vt))
}

//export goVRelease
func goVRelease(pVTab unsafe.Pointer, isDestroy int) *C.char {
	vt := lookupHandle(uintptr(pVTab)).(*sqliteVTab)
	var err error
	if isDestroy == 1 {
		err = vt.vTab.Destroy()
	} else {
		err = vt.vTab.Disconnect()
	}
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVOpen
func goVOpen(pVTab unsafe.Pointer, pzErr **C.char) C.uintptr_t {
	vt := lookupHandle(uintptr(pVTab)).(*sqliteVTab)
	vTabCursor, err := vt.vTab.Open()
	if err != nil {
		*pzErr = mPrintf("%s", err.Error())
		return 0
	}
	vtc := sqliteVTabCursor{vt, vTabCursor}
	*pzErr = nil
	return C.uintptr_t(newHandle(vt.module.c, &vtc))
}

//export goVBestIndex
func goVBestIndex(pVTab unsafe.Pointer, icp unsafe.Pointer) *C.char {
	vt := lookupHandle(uintptr(pVTab)).(*sqliteVTab)
	info := (*C.sqlite3_index_info)(icp)
	csts := constraints(info)
	res, err := vt.vTab.BestIndex(csts, orderBys(info))
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	if len(res.Used) != len(csts) {
		return mPrintf("Result.Used != expected value", "")
	}

	// Get a pointer to constraint_usage struct so we can update in place.
	l := info.nConstraint
	s := (*[1 << 30]C.struct_sqlite3_index_constraint_usage)(unsafe.Pointer(info.aConstraintUsage))[:l:l]
	index := 1
	for i := C.int(0); i < info.nConstraint; i++ {
		if res.Used[i] {
			s[i].argvIndex = C.int(index)
			s[i].omit = C.uchar(1)
			index++
		}
	}

	info.idxNum = C.int(res.IdxNum)
	idxStr := C.CString(res.IdxStr)
	defer C.free(unsafe.Pointer(idxStr))
	info.idxStr = idxStr
	info.needToFreeIdxStr = C.int(0)
	if res.AlreadyOrdered {
		info.orderByConsumed = C.int(1)
	}
	info.estimatedCost = C.double(res.EstimatedCost)
	info.estimatedRows = C.sqlite3_int64(res.EstimatedRows)

	return nil
}

//export goVClose
func goVClose(pCursor unsafe.Pointer) *C.char {
	vtc := lookupHandle(uintptr(pCursor)).(*sqliteVTabCursor)
	err := vtc.vTabCursor.Close()
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goMDestroy
func goMDestroy(pClientData unsafe.Pointer) {
	m := lookupHandle(uintptr(pClientData)).(*sqliteModule)
	m.module.DestroyModule()
}

//export goVFilter
func goVFilter(pCursor unsafe.Pointer, idxNum int, idxName *C.char, argc int, argv **C.sqlite3_value) *C.char {
	vtc := lookupHandle(uintptr(pCursor)).(*sqliteVTabCursor)
	args := (*[(math.MaxInt32 - 1) / unsafe.Sizeof((*C.sqlite3_value)(nil))]*C.sqlite3_value)(unsafe.Pointer(argv))[:argc:argc]
	vals := make([]interface{}, 0, argc)
	for _, v := range args {
		conv, err := callbackArgGeneric(v)
		if err != nil {
			return mPrintf("%s", err.Error())
		}
		vals = append(vals, conv.Interface())
	}
	err := vtc.vTabCursor.Filter(idxNum, C.GoString(idxName), vals)
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVNext
func goVNext(pCursor unsafe.Pointer) *C.char {
	vtc := lookupHandle(uintptr(pCursor)).(*sqliteVTabCursor)
	err := vtc.vTabCursor.Next()
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVEof
func goVEof(pCursor unsafe.Pointer) C.int {
	vtc := lookupHandle(uintptr(pCursor)).(*sqliteVTabCursor)
	err := vtc.vTabCursor.EOF()
	if err {
		return 1
	}
	return 0
}

//export goVColumn
func goVColumn(pCursor, cp unsafe.Pointer, col int) *C.char {
	vtc := lookupHandle(uintptr(pCursor)).(*sqliteVTabCursor)
	c := (*SQLiteContext)(cp)
	err := vtc.vTabCursor.Column(c, col)
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	return nil
}

//export goVRowid
func goVRowid(pCursor unsafe.Pointer, pRowid *C.sqlite3_int64) *C.char {
	vtc := lookupHandle(uintptr(pCursor)).(*sqliteVTabCursor)
	rowid, err := vtc.vTabCursor.Rowid()
	if err != nil {
		return mPrintf("%s", err.Error())
	}
	*pRowid = C.sqlite3_int64(rowid)
	return nil
}

// Module is a "virtual table module", it defines the implementation of a
// virtual tables. See: http://sqlite.org/c3ref/module.html
type Module interface {
	// http://sqlite.org/vtab.html#xcreate
	Create(c *SQLiteConn, args []string) (VTab, error)
	// http://sqlite.org/vtab.html#xconnect
	Connect(c *SQLiteConn, args []string) (VTab, error)
	// http://sqlite.org/c3ref/create_module.html
	DestroyModule()
}

// VTab describes a particular instance of the virtual table.
// See: http://sqlite.org/c3ref/vtab.html
type VTab interface {
	// http://sqlite.org/vtab.html#xbestindex
	BestIndex([]InfoConstraint, []InfoOrderBy) (*IndexResult, error)
	// http://sqlite.org/vtab.html#xdisconnect
	Disconnect() error
	// http://sqlite.org/vtab.html#sqlite3_module.xDestroy
	Destroy() error
	// http://sqlite.org/vtab.html#xopen
	Open() (VTabCursor, error)
}

// VTabCursor describes cursors that point into the virtual table and are used
// to loop through the virtual table. See: http://sqlite.org/c3ref/vtab_cursor.html
type VTabCursor interface {
	// http://sqlite.org/vtab.html#xclose
	Close() error
	// http://sqlite.org/vtab.html#xfilter
	Filter(idxNum int, idxStr string, vals []interface{}) error
	// http://sqlite.org/vtab.html#xnext
	Next() error
	// http://sqlite.org/vtab.html#xeof
	EOF() bool
	// http://sqlite.org/vtab.html#xcolumn
	Column(c *SQLiteContext, col int) error
	// http://sqlite.org/vtab.html#xrowid
	Rowid() (int64, error)
}

// DeclareVTab declares the Schema of a virtual table.
// See: http://sqlite.org/c3ref/declare_vtab.html
func (c *SQLiteConn) DeclareVTab(sql string) error {
	zSQL := C.CString(sql)
	defer C.free(unsafe.Pointer(zSQL))
	rv := C.sqlite3_declare_vtab(c.db, zSQL)
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	return nil
}

// CreateModule registers a virtual table implementation.
// See: http://sqlite.org/c3ref/create_module.html
func (c *SQLiteConn) CreateModule(moduleName string, module Module) error {
	mname := C.CString(moduleName)
	defer C.free(unsafe.Pointer(mname))
	udm := sqliteModule{c, moduleName, module}
	rv := C.goSqlite3CreateModule(c.db, mname, C.uintptr_t(newHandle(c, &udm)))
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	return nil
}
