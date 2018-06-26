// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build go1.10

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"testing"
)

func TestOpenConnector(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	drv := &SQLiteDriver{}

	connector, err := drv.OpenConnector(tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := connector.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("Failed to create connection to database")
	}
	defer conn.Close()

	stmt, err := conn.Prepare("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec([]driver.Value{}); err != nil {
		t.Fatalf("Failed to exec statement: %s", err)
	}

	// Verify database has been created
	if _, err := os.Stat(tempFilename); os.IsNotExist(err) {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}
}

func TestOpenConnectorInvalidDSN(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	drv := &SQLiteDriver{}

	_, err := drv.OpenConnector(fmt.Sprintf("file:%s?%35%2%%43?test=false", tempFilename))
	if err == nil {
		t.Fatal("Opened connector while error was expected")
	}
}

func TestOpenDB(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	cfg := NewConfig()
	cfg.Database = tempFilename

	// OpenDB
	db := sql.OpenDB(cfg)
	defer db.Close()

	_, err := db.Exec("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}

	if _, err := os.Stat(tempFilename); os.IsNotExist(err) {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}
}

func TestOpenDBReadOnly(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	cfg := NewConfig()
	cfg.Database = tempFilename

	// OpenDB
	db1 := sql.OpenDB(cfg)
	defer db1.Close()
	db1.Exec("CREATE TABLE test (x int, y float)")

	cfg.Mode = ModeReadOnly
	db2 := sql.OpenDB(cfg)
	defer db2.Close()

	_, err := db2.Exec("INSERT INTO test VALUES (1, 3.14)")
	if err == nil {
		t.Fatal("didn't expect INSERT into read-only database to work")
	}
}
