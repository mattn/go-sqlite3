// +build !go1.21
// +build libsqlite3,!sqlite_serialize

package sqlite3

import (
	"errors"
)

import "C"

func (c *SQLiteConn) DeserializeEmbedded(b []byte, schema string) error {
	return errors.New("sqlite3: DeserializeEmbedded requires go1.21+ and the sqlite_serialize build tag when using the libsqlite3 build tag")
}
