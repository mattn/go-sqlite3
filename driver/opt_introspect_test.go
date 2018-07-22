// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build sqlite_introspect

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

func TestIntrospectFunctionList(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("PRAGMA function_list;")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var s string
		var s2 string
		err := rows.Scan(&s, &s2)
		if err != nil {
			t.Fatal(err)
		}
		list = append(list, s)
	}

	if len(list) == 0 {
		t.Fatal("introspect: No Function List Available")
	}
}

func TestIntrospectModuleList(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("PRAGMA module_list;")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			t.Fatal(err)
		}
		list = append(list, s)
	}

	if len(list) == 0 {
		t.Fatal("introspect: No Module List Available")
	}
}

func TestIntrospectPRAGMAList(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("PRAGMA pragma_list;")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			t.Fatal(err)
		}
		list = append(list, s)
	}

	if len(list) == 0 {
		t.Fatal("introspect: No PRAGMA List Available")
	}
}
