// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build sqlite_stat4

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

func TestStat4(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", "file:"+tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	_, err = db.Exec("create index idx_id on foo(id)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := db.Begin()
	for i := 0; i < 10000; i++ {
		tx.Exec("insert into foo(id) VALUES(?)", i)
	}
	tx.Commit()

	if _, err := db.Exec("ANALYZE"); err != nil {
		t.Fatal(err)
	}

	exists := 0
	if err := db.QueryRow("select count(type) from sqlite_master where type = 'table' and name = 'sqlite_stat4';").Scan(&exists); err != nil {
		t.Fatal(err)
	}

	if exists != 1 {
		t.Fatal("Failed to enable STAT4")
	}
}
