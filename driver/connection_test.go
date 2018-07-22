// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConnGetFileName(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	driverName := fmt.Sprintf("sqlite3-%s", t.Name())

	var driverConn *SQLiteConn
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			driverConn = conn
			return nil
		},
	})

	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	abs, err := filepath.Abs(tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	if fh := driverConn.GetFilename(""); fh != abs {
		t.Fatalf("GetFileName failed; expected: %s, got: %s", fh, abs)
	}
}

func TestConnAutoCommit(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	driverName := fmt.Sprintf("sqlite3-%s", t.Name())

	var driverConn *SQLiteConn
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			driverConn = conn
			return nil
		},
	})

	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	if !driverConn.AutoCommit() {
		t.Fatal("autocommit was expected to be true")
	}
}
