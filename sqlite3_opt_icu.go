// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_icu icu

package sqlite3

/*
#cgo LDFLAGS: -licuuc
#cgo CFLAGS: -DSQLITE_ENABLE_ICU
#cgo linux LDFLAGS: -licui18n
#cgo darwin CFLAGS: -I/usr/local/opt/icu4c/include
#cgo darwin LDFLAGS: -L/usr/local/opt/icu4c/lib -licui18n
#cgo openbsd LDFLAGS: -lsqlite3 -licui18n
#cgo windows LDFLAGS: -licuin
*/
import "C"
