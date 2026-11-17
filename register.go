// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

import "database/sql"

// This variable can be replaced with -ldflags like below:
// go build -ldflags="-X 'github.com/mattn/go-sqlite3.driverName=my-sqlite3'"
var driverName = "sqlite3"

func init() {
	if driverName != "" {
		sql.Register(driverName, &SQLiteDriver{})
	}
}
