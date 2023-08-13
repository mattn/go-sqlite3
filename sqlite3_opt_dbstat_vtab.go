// +build sqlite_dbstat_vtab dbstat_vtab

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_DBSTAT_VTAB=1
#cgo LDFLAGS: -lm
*/
import "C"
