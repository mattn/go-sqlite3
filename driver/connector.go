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
	_ driver.Connector = (*Connector)(nil)
)

// Connector is a driver in a fixed configuration.
type Connector struct {
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c Connector) Connect(ctx context.Context) (driver.Conn, error) {
	return nil, nil
}

// Driver implements driver.Connector interface.
// Driver returns &MySQLDriver{}.
func (c Connector) Driver() driver.Driver {
	return nil
}
