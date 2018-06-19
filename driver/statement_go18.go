// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build go1.8

package sqlite3

import (
	"context"
	"database/sql/driver"
)

var (
	_ driver.StmtExecContext  = (*SQLiteStmt)(nil)
	_ driver.StmtQueryContext = (*SQLiteStmt)(nil)
)

// QueryContext implement QueryerContext.
func (s *SQLiteStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	list := make([]namedValue, len(args))
	for i, nv := range args {
		list[i] = namedValue(nv)
	}
	return s.query(ctx, list)
}

// ExecContext implement ExecerContext.
func (s *SQLiteStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	list := make([]namedValue, len(args))
	for i, nv := range args {
		list[i] = namedValue(nv)
	}
	return s.exec(ctx, list)
}
