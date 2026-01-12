//go:build go1.21
// +build go1.21

// The unsafe.StringData function was made available in Go 1.20 but it
// was not until Go 1.21 that Go was changed to interpret the Go version
// in go.mod (1.19 as of writing this) as the minimum version required
// instead of the exact version.
//
// See: https://github.com/golang/go/issues/59033

package sqlite3

import "unsafe"

// stringData is a safe version of unsafe.StringData that handles empty strings.
func stringData(s string) *byte {
	if len(s) != 0 {
		return unsafe.StringData(s)
	}
	// The return value of unsafe.StringData
	// is unspecified if the string is empty.
	return &placeHolder[0]
}
