// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !cgo

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"errors"
)

// Implementation Enforcer
var (
	_ driver.Driver = (*SQLiteDriver)(nil)
)

func init() {
	sql.Register("sqlite3", &SQLiteDriver{})
}

type SQLiteDriver struct{}

var errCgoDisabled = errors.New("binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to be enabled.")

func (d *SQLiteDriver) Open(dsn string) (driver.Conn, error) {
	return nil, errCgoDisabled
}
