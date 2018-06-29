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
	"reflect"
	"testing"
	"time"
)

func TestResultColumnTypeInteger(t *testing.T) {
	testColumnType(t, "INTEGER", 1, reflect.TypeOf(int64(0)))
}

func TestResultColumnTypeText(t *testing.T) {
	testColumnType(t, "TEXT", "FooBar", reflect.TypeOf(""))
}

func TestResultColumnTypeBLOB(t *testing.T) {
	testColumnType(t, "BLOB", []byte{'\x20'}, reflect.TypeOf([]byte{}))
}

func TestResultColumnTypeFloat(t *testing.T) {
	testColumnType(t, "REAL", float64(1.3), reflect.TypeOf(float64(0)))
}

func TestResultColumnTypeBoolean(t *testing.T) {
	testColumnType(t, "BOOLEAN", true, reflect.TypeOf(true))
}

func TestResultColumnTypeDateTime(t *testing.T) {
	testColumnType(t, "DATETIME", time.Now().Unix(), reflect.TypeOf(time.Time{}))
}

func TestResultColumnTypeNULL(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec("CREATE TABLE type (t TEXT);"); err != nil {
		t.Fatal(err)
	}

	ins, err := db.Prepare("INSERT INTO type (t) VALUES(?)")
	if err != nil {
		t.Fatalf("Prepare Failed: %v", err)
	}

	if ins.Exec(nil); err != nil {
		t.Fatalf("Insert Failed: %v", err)
	}

	rows, err := db.Query("SELECT t FROM type LIMIT 1;")
	if err != nil {
		t.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		types, err := rows.ColumnTypes()
		if err != nil {
			t.Fatal(err)
		}

		if types[0].DatabaseTypeName() != "TEXT" {
			t.Fatalf("Invalid column type; expected: %s, got: %s", "TEXT", types[0].DatabaseTypeName())
		}

		if v := types[0].ScanType(); v != reflect.TypeOf(nil) {
			t.Fatalf("Wrong scan type returned, expected: %v; got: %v", reflect.TypeOf(nil), v)
		}
	}
}

func testColumnType(t *testing.T, dtype string, def interface{}, rtype reflect.Type) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE TABLE type (t %s NOT NULL);", dtype)); err != nil {
		t.Fatal(err)
	}

	ins, err := db.Prepare("INSERT INTO type (t) VALUES(?)")
	if err != nil {
		t.Fatalf("Prepare Failed: %v", err)
	}

	if ins.Exec(def); err != nil {
		t.Fatalf("Insert Failed: %v", err)
	}

	rows, err := db.Query("SELECT t FROM type LIMIT 1;")
	if err != nil {
		t.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		types, err := rows.ColumnTypes()
		if err != nil {
			t.Fatal(err)
		}

		if types[0].DatabaseTypeName() != dtype {
			t.Fatalf("Invalid column type; expected: %s, got: %s", dtype, types[0].DatabaseTypeName())
		}

		if v := types[0].ScanType(); v != rtype {
			t.Fatalf("Wrong scan type returned, expected: %v; got: %v", rtype, v)
		}
	}
}
