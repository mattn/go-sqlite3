package sqlite3

import "unsafe"

func WalHookInternalDelete(conn *SQLiteConn) {
	delete(walHooks, uintptr(unsafe.Pointer(conn.db)))
}
