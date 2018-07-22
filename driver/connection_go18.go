// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build go1.8

package sqlite3

import (
	"database/sql/driver"
	"errors"

	"context"
)

// Enforce implementation
var (
	_ driver.ConnBeginTx        = (*SQLiteConn)(nil)
	_ driver.ConnPrepareContext = (*SQLiteConn)(nil)
	_ driver.ExecerContext      = (*SQLiteConn)(nil)
	_ driver.QueryerContext     = (*SQLiteConn)(nil)
	_ driver.Pinger             = (*SQLiteConn)(nil)
)

// Ping implement Pinger.
func (c *SQLiteConn) Ping(ctx context.Context) error {
	if c.db == nil {
		return errors.New("connection was closed")
	}

	return nil
}

// QueryContext implement QueryerContext.
func (c *SQLiteConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	list := make([]namedValue, len(args))
	for i, nv := range args {
		list[i] = namedValue(nv)
	}
	return c.query(ctx, query, list)
}

// ExecContext implement ExecerContext.
func (c *SQLiteConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	list := make([]namedValue, len(args))
	for i, nv := range args {
		list[i] = namedValue(nv)
	}
	return c.exec(ctx, query, list)
}

// PrepareContext implement ConnPrepareContext.
func (c *SQLiteConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return c.prepare(ctx, query)
}

// BeginTx implement ConnBeginTx.
func (c *SQLiteConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.begin(ctx)
}
