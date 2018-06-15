// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !cgo
// +build go1.10

package sqlite3

import (
	"database/sql/driver"
)

// Implementation Enforcer
var (
	_ driver.DriverContext = (*SQLiteDriver)(nil)
)

func (d *SQLiteDriver) OpenConnector(name string) (driver.Connector, error) {
	return nil, errCgoDisabled
}
