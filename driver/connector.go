// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build go1.10

package sqlite3

import (
	"context"
	"database/sql/driver"
)

var (
	_ driver.Connector = (*Config)(nil)
)

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c *Config) Connect(ctx context.Context) (driver.Conn, error) {
	return c.createConnection()
}

// Driver implements driver.Connector interface.
// Driver returns &SQLiteDriver{}.
func (c *Config) Driver() driver.Driver {
	return &SQLiteDriver{
		Config:      c,
		Extensions:  c.Extensions,
		ConnectHook: c.ConnectHook,
	}
}
