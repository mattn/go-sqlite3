// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build go1.8

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestNamedParams(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	create table foo (id integer, name text, extra text);
	`)
	if err != nil {
		t.Error("Failed to call db.Query:", err)
	}

	_, err = db.Exec(`insert into foo(id, name, extra) values(:id, :name, :name)`, sql.Param(":name", "foo"), sql.Param(":id", 1))
	if err != nil {
		t.Error("Failed to call db.Exec:", err)
	}

	row := db.QueryRow(`select id, extra from foo where id = :id and extra = :extra`, sql.Param(":id", 1), sql.Param(":extra", "foo"))
	if row == nil {
		t.Error("Failed to call db.QueryRow")
	}
	var id int
	var extra string
	err = row.Scan(&id, &extra)
	if err != nil {
		t.Error("Failed to db.Scan:", err)
	}
	if id != 1 || extra != "foo" {
		t.Error("Failed to db.QueryRow: not matched results")
	}
}

func TestMultipleResultSet(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	create table foo (id integer, name text);
	`)
	if err != nil {
		t.Error("Failed to call db.Query:", err)
	}

	for i := 0; i < 100; i++ {
		_, err = db.Exec(`insert into foo(id, name) values(?, ?)`, i+1, fmt.Sprintf("foo%03d", i+1))
		if err != nil {
			t.Error("Failed to call db.Exec:", err)
		}
	}

	rows, err := db.Query(`
	select id, name from foo where id < :id1;
	select id, name from foo where id = :id2;
	select id, name from foo where id > :id3;
	`,
		sql.Param(":id1", 3),
		sql.Param(":id2", 50),
		sql.Param(":id3", 98),
	)
	if err != nil {
		t.Error("Failed to call db.Query:", err)
	}

	var id int
	var extra string

	for {
		for rows.Next() {
			err = rows.Scan(&id, &extra)
			if err != nil {
				t.Error("Failed to db.Scan:", err)
			}
			if id != 1 || extra != "foo" {
				t.Error("Failed to db.QueryRow: not matched results")
			}
		}
		if !rows.NextResultSet() {
			break
		}
	}
}
