// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build libsqlite3

package sqlite3

/*
#cgo CFLAGS: -DUSE_LIBSQLITE3
#cgo !darwin LDFLAGS: -lsqlite3
#cgo darwin LDFLAGS: -L/opt/homebrew/opt/sqlite/lib -L/usr/local/opt/sqlite/lib -lsqlite3
#cgo darwin CFLAGS: -I/opt/homebrew/opt/sqlite/include -I/usr/local/opt/sqlite/include
*/
import "C"
