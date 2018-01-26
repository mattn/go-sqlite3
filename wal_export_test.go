package sqlite3

func WalHookInternalDelete(conn *SQLiteConn) {
	delete(walHooks, conn.db)
}
