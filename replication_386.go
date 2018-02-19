// +build linux,386

package sqlite3

import "unsafe"

func unsafePointerToSlice(pList unsafe.Pointer, n int) []ReplicationPage {
	return (*[1 << 16]ReplicationPage)(pList)[:n:n]
}
