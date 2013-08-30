// +build !windows

package sqlite3

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -v
#cgo linux LDFLAGS: -ldl
*/
import "C"
