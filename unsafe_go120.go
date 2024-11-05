//go:build !go1.21
// +build !go1.21

package sqlite3

import "unsafe"

// stringData is a safe version of unsafe.StringData that handles empty strings.
func stringData(s string) *byte {
	if len(s) != 0 {
		b := *(*[]byte)(unsafe.Pointer(&s))
		return &b[0]
	}
	// The return value of unsafe.StringData
	// is unspecified if the string is empty.
	return &placeHolder[0]
}
