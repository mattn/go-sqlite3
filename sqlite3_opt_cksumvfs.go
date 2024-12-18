//go:build cksumvfs
// +build cksumvfs

package sqlite3

//extern int sqlite3_register_cksumvfs(const char*);
import "C"

func InitCksumVFS() {
	C.sqlite3_register_cksumvfs(nil)
}
