// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>

#ifndef SQLITE_DETERMINISTIC
# define SQLITE_DETERMINISTIC 0
#endif

void callbackTrampoline(sqlite3_context*, int, sqlite3_value**);
void stepTrampoline(sqlite3_context*, int, sqlite3_value**);
void doneTrampoline(sqlite3_context*);

int compareTrampoline(void*, int, char*, int, char*);

int _sqlite3_create_function(
  sqlite3 *db,
  const char *zFunctionName,
  int nArg,
  int eTextRep,
  uintptr_t pApp,
  void (*xFunc)(sqlite3_context*,int,sqlite3_value**),
  void (*xStep)(sqlite3_context*,int,sqlite3_value**),
  void (*xFinal)(sqlite3_context*)
) {
  return sqlite3_create_function(db, zFunctionName, nArg, eTextRep, (void*) pApp, xFunc, xStep, xFinal);
}
*/
import "C"
import (
	"errors"
	"reflect"
	"unsafe"
)

func sqlite3CreateFunction(db *C.sqlite3, zFunctionName *C.char, nArg C.int, eTextRep C.int, pApp uintptr, xFunc unsafe.Pointer, xStep unsafe.Pointer, xFinal unsafe.Pointer) C.int {
	return C._sqlite3_create_function(db, zFunctionName, nArg, eTextRep, C.uintptr_t(pApp), (*[0]byte)(xFunc), (*[0]byte)(xStep), (*[0]byte)(xFinal))
}

// RegisterCollation makes a Go function available as a collation.
//
// cmp receives two UTF-8 strings, a and b. The result should be 0 if
// a==b, -1 if a < b, and +1 if a > b.
//
// cmp must always return the same result given the same
// inputs. Additionally, it must have the following properties for all
// strings A, B and C: if A==B then B==A; if A==B and B==C then A==C;
// if A<B then B>A; if A<B and B<C then A<C.
//
// If cmp does not obey these constraints, sqlite3's behavior is
// undefined when the collation is used.
func (c *SQLiteConn) RegisterCollation(name string, cmp func(string, string) int) error {
	handle := newHandle(c, cmp)
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	rv := C.sqlite3_create_collation(c.db, cname, C.SQLITE_UTF8, unsafe.Pointer(handle), (*[0]byte)(unsafe.Pointer(C.compareTrampoline)))
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	return nil
}

// RegisterFunc makes a Go function available as a SQLite function.
//
// The Go function can have arguments of the following types: any
// numeric type except complex, bool, []byte, string and
// interface{}. interface{} arguments are given the direct translation
// of the SQLite data type: int64 for INTEGER, float64 for FLOAT,
// []byte for BLOB, string for TEXT.
//
// The function can additionally be variadic, as long as the type of
// the variadic argument is one of the above.
//
// If pure is true. SQLite will assume that the function's return
// value depends only on its inputs, and make more aggressive
// optimizations in its queries.
//
// See _example/go_custom_funcs for a detailed example.
func (c *SQLiteConn) RegisterFunc(name string, impl interface{}, pure bool) error {
	var fi functionInfo
	fi.f = reflect.ValueOf(impl)
	t := fi.f.Type()
	if t.Kind() != reflect.Func {
		return errors.New("non-function passed to RegisterFunc")
	}
	if t.NumOut() != 1 && t.NumOut() != 2 {
		return errors.New("SQLite functions must return 1 or 2 values")
	}
	if t.NumOut() == 2 && !t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("second return value of SQLite function must be error")
	}

	numArgs := t.NumIn()
	if t.IsVariadic() {
		numArgs--
	}

	for i := 0; i < numArgs; i++ {
		conv, err := callbackArg(t.In(i))
		if err != nil {
			return err
		}
		fi.argConverters = append(fi.argConverters, conv)
	}

	if t.IsVariadic() {
		conv, err := callbackArg(t.In(numArgs).Elem())
		if err != nil {
			return err
		}
		fi.variadicConverter = conv
		// Pass -1 to sqlite so that it allows any number of
		// arguments. The call helper verifies that the minimum number
		// of arguments is present for variadic functions.
		numArgs = -1
	}

	conv, err := callbackRet(t.Out(0))
	if err != nil {
		return err
	}
	fi.retConverter = conv

	// fi must outlast the database connection, or we'll have dangling pointers.
	c.funcs = append(c.funcs, &fi)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	opts := C.SQLITE_UTF8
	if pure {
		opts |= C.SQLITE_DETERMINISTIC
	}
	rv := sqlite3CreateFunction(c.db, cname, C.int(numArgs), C.int(opts), newHandle(c, &fi), C.callbackTrampoline, nil, nil)
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	return nil
}

// RegisterAggregator makes a Go type available as a SQLite aggregation function.
//
// Because aggregation is incremental, it's implemented in Go with a
// type that has 2 methods: func Step(values) accumulates one row of
// data into the accumulator, and func Done() ret finalizes and
// returns the aggregate value. "values" and "ret" may be any type
// supported by RegisterFunc.
//
// RegisterAggregator takes as implementation a constructor function
// that constructs an instance of the aggregator type each time an
// aggregation begins. The constructor must return a pointer to a
// type, or an interface that implements Step() and Done().
//
// The constructor function and the Step/Done methods may optionally
// return an error in addition to their other return values.
//
// See _example/go_custom_funcs for a detailed example.
func (c *SQLiteConn) RegisterAggregator(name string, impl interface{}, pure bool) error {
	var ai aggInfo
	ai.constructor = reflect.ValueOf(impl)
	t := ai.constructor.Type()
	if t.Kind() != reflect.Func {
		return errors.New("non-function passed to RegisterAggregator")
	}
	if t.NumOut() != 1 && t.NumOut() != 2 {
		return errors.New("SQLite aggregator constructors must return 1 or 2 values")
	}
	if t.NumOut() == 2 && !t.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("second return value of SQLite function must be error")
	}
	if t.NumIn() != 0 {
		return errors.New("SQLite aggregator constructors must not have arguments")
	}

	agg := t.Out(0)
	switch agg.Kind() {
	case reflect.Ptr, reflect.Interface:
	default:
		return errors.New("SQlite aggregator constructor must return a pointer object")
	}
	stepFn, found := agg.MethodByName("Step")
	if !found {
		return errors.New("SQlite aggregator doesn't have a Step() function")
	}
	step := stepFn.Type
	if step.NumOut() != 0 && step.NumOut() != 1 {
		return errors.New("SQlite aggregator Step() function must return 0 or 1 values")
	}
	if step.NumOut() == 1 && !step.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("type of SQlite aggregator Step() return value must be error")
	}

	stepNArgs := step.NumIn()
	start := 0
	if agg.Kind() == reflect.Ptr {
		// Skip over the method receiver
		stepNArgs--
		start++
	}
	if step.IsVariadic() {
		stepNArgs--
	}
	for i := start; i < start+stepNArgs; i++ {
		conv, err := callbackArg(step.In(i))
		if err != nil {
			return err
		}
		ai.stepArgConverters = append(ai.stepArgConverters, conv)
	}
	if step.IsVariadic() {
		conv, err := callbackArg(t.In(start + stepNArgs).Elem())
		if err != nil {
			return err
		}
		ai.stepVariadicConverter = conv
		// Pass -1 to sqlite so that it allows any number of
		// arguments. The call helper verifies that the minimum number
		// of arguments is present for variadic functions.
		stepNArgs = -1
	}

	doneFn, found := agg.MethodByName("Done")
	if !found {
		return errors.New("SQlite aggregator doesn't have a Done() function")
	}
	done := doneFn.Type
	doneNArgs := done.NumIn()
	if agg.Kind() == reflect.Ptr {
		// Skip over the method receiver
		doneNArgs--
	}
	if doneNArgs != 0 {
		return errors.New("SQLite aggregator Done() function must have no arguments")
	}
	if done.NumOut() != 1 && done.NumOut() != 2 {
		return errors.New("SQLite aggregator Done() function must return 1 or 2 values")
	}
	if done.NumOut() == 2 && !done.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("second return value of SQLite aggregator Done() function must be error")
	}

	conv, err := callbackRet(done.Out(0))
	if err != nil {
		return err
	}
	ai.doneRetConverter = conv
	ai.active = make(map[int64]reflect.Value)
	ai.next = 1

	// ai must outlast the database connection, or we'll have dangling pointers.
	c.aggregators = append(c.aggregators, &ai)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	opts := C.SQLITE_UTF8
	if pure {
		opts |= C.SQLITE_DETERMINISTIC
	}
	rv := sqlite3CreateFunction(c.db, cname, C.int(stepNArgs), C.int(opts), newHandle(c, &ai), nil, C.stepTrampoline, C.doneTrampoline)
	if rv != C.SQLITE_OK {
		return c.lastError()
	}
	return nil
}
