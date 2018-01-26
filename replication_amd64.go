// +build linux,amd64

package sqlite3

import "unsafe"

func unsafePointerToSlice(pList unsafe.Pointer, n int) []ReplicationPage {
	return (*[1 << 30]ReplicationPage)(pList)[:n:n]
}
