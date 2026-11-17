// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo

package sqlite3

import (
	"context"
	"database/sql/driver"
)

// Ping implement Pinger.
func (c *SQLiteConn) Ping(ctx context.Context) error {
	if c.db == nil {
		// must be ErrBadConn for sql to close the database
		return driver.ErrBadConn
	}
	return nil
}

// ResetSession implements driver.SessionResetter.
// SQLite statements and transactions are reset when they finish, so a healthy
// connection needs no additional cleanup before reuse.
func (c *SQLiteConn) ResetSession(ctx context.Context) error {
	if !c.IsValid() {
		return driver.ErrBadConn
	}
	return nil
}

// IsValid reports whether the connection can be reused.
func (c *SQLiteConn) IsValid() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db != nil && !c.rollbackFailed
}

// QueryContext implement QueryerContext.
func (c *SQLiteConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.query(ctx, query, args)
}

// ExecContext implement ExecerContext.
func (c *SQLiteConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.exec(ctx, query, args)
}

// PrepareContext implement ConnPrepareContext.
func (c *SQLiteConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return c.prepare(ctx, query)
}

// BeginTx implement ConnBeginTx.
func (c *SQLiteConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.begin(ctx)
}

// QueryContext implement QueryerContext.
func (s *SQLiteStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return s.query(ctx, args)
}

// ExecContext implement ExecerContext.
func (s *SQLiteStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return s.exec(ctx, args)
}
