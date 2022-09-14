// +build libsqlite3,!sqlite_serialize
// +build !sqlite_serialize

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_OMIT_DESERIALIZE
*/
import "C"

import (
	"fmt"
)

func (c *SQLiteConn) Serialize(schema string) ([]byte, error)  {
	return nil, errors.New("sqlite3: Serialize requires the sqlite_serialize build tag when using the libsqlite3 build tag")
}

func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	return fmt.Errorf("deserialize function not available")
}
