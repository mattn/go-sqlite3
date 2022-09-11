// +build sqlite_omit_deserialize

package sqlite3

func (c *SQLiteConn) Serialize(schema string) []byte {
	return nil
}

func (c *SQLiteConn) Deserialize(b []byte, schema string) error {
	return fmt.Errorf("deserialize function not available")
}
