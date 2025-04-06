// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build libsqlite3
// +build libsqlite3

package sqlite3

/*
#cgo CFLAGS: -DUSE_LIBSQLITE3
#cgo linux LDFLAGS: -lsqlcipher        // Use system SQLCipher instead of SQLite
#cgo darwin,amd64 LDFLAGS: -L/usr/local/opt/sqlcipher/lib -lsqlcipher
#cgo darwin,amd64 CFLAGS:  -I/usr/local/opt/sqlcipher/include
#cgo darwin,arm64 LDFLAGS: -L/opt/homebrew/opt/sqlcipher/lib -lsqlcipher
#cgo darwin,arm64 CFLAGS:  -I/opt/homebrew/opt/sqlcipher/include
#cgo openbsd LDFLAGS: -lsqlcipher
#cgo solaris LDFLAGS: -lsqlcipher
#cgo windows LDFLAGS: -lsqlcipher
#cgo zos LDFLAGS: -lsqlcipher
#include <sqlite3.h>
#include <stdlib.h>
*/
import "C"
