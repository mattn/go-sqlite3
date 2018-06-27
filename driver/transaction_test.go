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
	"testing"
)

func TestTransaction(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id) VALUES(1)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	rows, err := tx.Query("SELECT id from foo")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatal("Failed to rollback transaction:", err)
	}

	if rows.Next() {
		t.Fatal("Unable to query results:", err)
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id) VALUES(1)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal("Failed to commit transaction:", err)
	}

	rows, err = tx.Query("SELECT id from foo")
	if err == nil {
		t.Fatal("Expected failure to query")
	}
}

func TestTransactionConn(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	driverName := fmt.Sprintf("sqlite3.%s", t.Name())

	// The driver's connection will be needed in order to perform the backup.
	var dbConn *SQLiteConn
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			dbConn = conn
			return nil
		},
	})

	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if tx == nil {
		t.Fatal("Failed to create Transaction")
	}
}
