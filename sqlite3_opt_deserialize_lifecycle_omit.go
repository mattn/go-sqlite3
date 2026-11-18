//go:build libsqlite3 && !sqlite_serialize

package sqlite3

func (c *SQLiteConn) clearDeserializeBufLocked() {}
