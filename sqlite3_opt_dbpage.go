// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_dbpage
// +build sqlite_dbpage

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_DBPAGE_VTAB
#cgo LDFLAGS: -lm
*/
import "C"
