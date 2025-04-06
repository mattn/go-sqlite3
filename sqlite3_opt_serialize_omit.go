//go:build libsqlite3 && !sqlite_serialize
// +build libsqlite3,!sqlite_serialize

package sqlite3

import (
	"errors"
)

/*
#cgo CFLAGS: -DSQLITE_OMIT_DESERIALIZE
#cgo LDFLAGS: -lcrypto -lsqlcipher
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h" // Use amalgamation if enabled
#else
#include <sqlcipher/sqlite3.h> // Use system-provided SQLCipher
#endif
*/
import "C"

func (c *SQLiteConn) Serialize(schema string) ([]byte, error) {
	return nil, errors.New("sqlite3: Serialize requires the sqlite_serialize build tag when using the libsqlite3 build tag")
}

func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	return errors.New("sqlite3: Deserialize requires the sqlite_serialize build tag when using the libsqlite3 build tag")
}
