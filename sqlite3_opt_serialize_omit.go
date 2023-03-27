// +build libsqlite3 libsqlcipher
// +build !sqlite_serialize

package sqlite3

import (
	"errors"
)

/*
#cgo CFLAGS: -DSQLITE_OMIT_DESERIALIZE
*/
import "C"

func (c *SQLiteConn) Serialize(schema string) ([]byte, error) {
	return nil, errors.New("sqlite3: Serialize requires the sqlite_serialize build tag when using the libsqlite3 or libsqlcipher build tags")
}

func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	return errors.New("sqlite3: Deserialize requires the sqlite_serialize build tag when using the libsqlite3 or libsqlcipher build tags")
}
