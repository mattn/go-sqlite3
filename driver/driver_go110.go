// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build go1.10

package sqlite3

import (
	"database/sql/driver"
)

var (
	_ driver.DriverContext = (*SQLiteDriver)(nil)
)

// OpenConnector will call OpenConnector to obtain a Connector and then invoke
// that Connector's Conn method to obtain each needed connection,
// instead of invoking the Driver's Open method for each connection.
// The two-step sequence allows drivers to parse the name just once and also provides
// access to per-Conn contexts.
func (d *SQLiteDriver) OpenConnector(dsn string) (driver.Connector, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	cfg, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// Configure Extensions
	cfg.Extensions = d.Extensions

	// Configure ConnectHook
	cfg.ConnectHook = d.ConnectHook

	// Set Configuration
	d.Config = cfg

	return cfg, nil
}
