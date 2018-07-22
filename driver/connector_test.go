// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build go1.10

package sqlite3

import (
	"context"
	"testing"
)

func TestConnectorDriver(t *testing.T) {
	// Create default Config
	cfg := NewConfig()
	cfg.ConnectHook = func(conn *SQLiteConn) error {
		return nil
	}

	// Create Driver from Config
	drv := cfg.Driver()
	if drv.(*SQLiteDriver).ConnectHook == nil {
		t.Fatal("failed to created Driver from Config")
	}
}

func TestConnectorConnect(t *testing.T) {
	// Create default Config
	cfg := NewConfig()

	// Create Connection to database from Config
	conn, err := cfg.Connect(context.Background())
	if err != nil || conn == nil {
		t.Fatal("failed to create connection from Config")
	}
}
