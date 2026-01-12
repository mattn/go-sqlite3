// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_fts5 || fts5
// +build sqlite_fts5 fts5

package sqlite3

import _ "github.com/mattn/go-sqlite3/fts5"
