// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "go-sqlite3-test-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestOpen(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatalf("Failed to open database: %s", err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}

	if stat, err := os.Stat(tempFilename); err != nil || stat.IsDir() {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}

	tempFilename = TempFilename(t)
	defer os.Remove(tempFilename)

	// Open Driver Directly
	drv := &SQLiteDriver{}
	conn, err := drv.Open(tempFilename)
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

func TestOpenInvalidDSN(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	// Open Invalid DSN
	drv := &SQLiteDriver{}
	conn, err := drv.Open(fmt.Sprintf("%s?%35%2%%43?test=false", tempFilename))
	if err == nil {
		t.Fatal("Connection created while error was expected")
	}
	if conn != nil {
		t.Fatal("Conection created while error was expected")
	}
}

func TestOpenConfigDSN(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	cfg := NewConfig()
	cfg.Database = tempFilename

	db, err := sql.Open("sqlite3", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("Failed to open database: %s", err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}

	if _, err := os.Stat(tempFilename); os.IsNotExist(err) {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}

	// Test Open Empry Database location
	cfg.Database = ""

	db, err = sql.Open("sqlite3", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err == nil {
		t.Fatalf("Table created while error was expected")
	}
}

func TestInvalidConnectHook(t *testing.T) {
	driverName := "sqlite3_invalid_connecthook"
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return fmt.Errorf("ConnectHook Error")
		},
	})

	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err == nil {
		t.Fatalf("Table created while error was expected")
	}
}

func TestInvalidExtension(t *testing.T) {
	driverName := "sqlite3_invalid_extension"
	sql.Register(driverName, &SQLiteDriver{
		Extensions: []string{
			"invalid.extension",
		},
	})

	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err == nil {
		t.Fatalf("Table created while error was expected")
	}

	tempFilename = TempFilename(t)
	defer os.Remove(tempFilename)

	driverName = "sqlite3_conn_invalid_extension"
	var driverConn *SQLiteConn
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			driverConn = conn
			return nil
		},
	})

	db, err = sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("SELECT 1;")
	if err != nil {
		t.Fatalf("Failed to exec ping statement")
	}

	if err := driverConn.LoadExtension("invalid.extension", ""); err == nil {
		t.Fatal("Extension loaded while error was expected")
	}
}
