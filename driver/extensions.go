// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !sqlite_omit_load_extension !sqlite_disable_extensions

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func (c *SQLiteConn) loadExtensions(extensions []string) error {
	// Enable extension loading
	rv := C.sqlite3_enable_load_extension(c.db, 1)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}

	for _, extension := range extensions {
		cext := C.CString(extension)
		cerr := C.CString("")
		defer C.free(unsafe.Pointer(cext))
		defer C.free(unsafe.Pointer(cerr))
		rv = C.sqlite3_load_extension(c.db, cext, nil, &cerr)
		if rv != C.SQLITE_OK {
			// Disable Extension Loading
			rv = C.sqlite3_enable_load_extension(c.db, 0)
			if rv != C.SQLITE_OK {
				return errors.New("Failed to disable extension loading")
			}
			fmt.Println(">>>")
			fmt.Println(*cerr)
			fmt.Printf("%v\n", cerr)
			fmt.Printf("%v\n", *cerr)
			return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
		}
	}

	// Disable extension loading
	rv = C.sqlite3_enable_load_extension(c.db, 0)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}
	return nil
}

// LoadExtension load the sqlite3 extension.
func (c *SQLiteConn) LoadExtension(lib string, entry string) error {
	rv := C.sqlite3_enable_load_extension(c.db, 1)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}

	clib := C.CString(lib)
	defer C.free(unsafe.Pointer(clib))
	centry := C.CString(entry)
	defer C.free(unsafe.Pointer(centry))

	rv = C.sqlite3_load_extension(c.db, clib, centry, nil)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}

	rv = C.sqlite3_enable_load_extension(c.db, 0)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}

	return nil
}
