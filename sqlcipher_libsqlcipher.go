// Copyright (C) 2022 Jonathan Giannuzzi <jonathan@giannuzzi.me>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build libsqlcipher

package sqlite3

/*
#cgo CFLAGS: -DUSE_LIBSQLCIPHER
#cgo !darwin LDFLAGS: -lsqlcipher
#cgo darwin LDFLAGS: -L/opt/homebrew/lib -L/usr/local/lib -lsqlcipher
#cgo darwin CFLAGS: -I/opt/homebrew/include -I/usr/local/include
*/
import "C"
