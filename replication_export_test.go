package sqlite3

import (
	"unsafe"
)

// Expose the replicationBegin hook for testing internal details.
func ReplicationBeginHook(pArg unsafe.Pointer) {
	replicationBegin(pArg)
}

// Expose registerMethodsInstance for testing internal details.
func ReplicationRegisterMethodsInstance(conn *SQLiteConn) func() {
	registerMethodsInstance(conn, nil)
	return func() {
		unregisterMethodsInstance(conn)
	}
}
