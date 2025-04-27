// Copyright (C) 2020 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build !libsqlite3

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestDeleteWithLimit(t *testing.T) {
	t.Skip("re-enable once bindings are updated to allow delete with limit")

	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	var expected int64
	for _, val := range []int{123, 456, 789} {
		res, err := db.Exec(fmt.Sprintf("insert into foo(id) values(%d)", val))
		if err != nil {
			t.Fatal("Failed to insert record:", err)
		}
		expected, err = res.LastInsertId()
		if err != nil {
			t.Fatal("Failed to get LastInsertId:", err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			t.Fatal("Failed to get RowsAffected:", err)
		}
		if affected != 1 {
			t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
		}
	}

	res, err := db.Exec("delete from foo where id = 123")
	if err != nil {
		t.Fatal("Failed to delete record:", err)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	if expected != lastID {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastID)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		t.Fatal("Failed to get RowsAffected:", err)
	}
	if affected != 1 {
		t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
	}

	res, err = db.Exec("delete from foo order by id asc limit 1")
	if err != nil {
		t.Fatal("Failed to delete record:", err)
	}
	lastID, err = res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	if expected != lastID {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastID)
	}
	affected, err = res.RowsAffected()
	if err != nil {
		t.Fatal("Failed to get RowsAffected:", err)
	}
	if affected != 1 {
		t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Expected remaining row")
	}
	var val int
	if err := rows.Scan(&val); err != nil {
		t.Fatal("Unable to scan results:", err)
	}
	if val != 789 {
		t.Errorf("Expected value of last row to be %d, but got %d", 789, val)
	}

	if rows.Next() {
		t.Error("Fetched row but expected not rows")
	}
}
