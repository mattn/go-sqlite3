// +build sqlite_omit_deserialize

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_OMIT_DESERIALIZE
*/
import "C"

import (
	"fmt"
)

func (c *SQLiteConn) Serialize(schema string) ([]byte, error)  {
	return nil, fmt.Errorf("serialize function not available")
}

func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	return fmt.Errorf("deserialize function not available")
}
