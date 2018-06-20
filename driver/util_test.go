// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"testing"
)

func TestVersion(t *testing.T) {
	s, n, id := Version()
	if s == "" || n == 0 || id == "" {
		t.Errorf("Version failed %q, %d, %q\n", s, n, id)
	}
}
