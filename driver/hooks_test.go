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
)

func TestHooksUpdateAndTransaction(t *testing.T) {
	var events []string
	var commitHookReturn = 0

	sql.Register("sqlite3_UpdateHook", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			conn.RegisterCommitHook(func() int {
				events = append(events, "commit")
				return commitHookReturn
			})
			conn.RegisterRollbackHook(func() {
				events = append(events, "rollback")
			})
			conn.RegisterUpdateHook(func(op int, db string, table string, rowid int64) {
				events = append(events, fmt.Sprintf("update(op=%v db=%v table=%v rowid=%v)", op, db, table, rowid))
			})
			return nil
		},
	})
	db, err := sql.Open("sqlite3_UpdateHook", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	statements := []string{
		"create table foo (id integer primary key)",
		"insert into foo values (9)",
		"update foo set id = 99 where id = 9",
		"delete from foo where id = 99",
	}
	for _, statement := range statements {
		_, err = db.Exec(statement)
		if err != nil {
			t.Fatalf("Unable to prepare test data [%v]: %v", statement, err)
		}
	}

	commitHookReturn = 1
	_, err = db.Exec("insert into foo values (5)")
	if err == nil {
		t.Error("Commit hook failed to rollback transaction")
	}

	var expected = []string{
		"commit",
		fmt.Sprintf("update(op=%v db=main table=foo rowid=9)", SQLITE_INSERT),
		"commit",
		fmt.Sprintf("update(op=%v db=main table=foo rowid=99)", SQLITE_UPDATE),
		"commit",
		fmt.Sprintf("update(op=%v db=main table=foo rowid=99)", SQLITE_DELETE),
		"commit",
		fmt.Sprintf("update(op=%v db=main table=foo rowid=5)", SQLITE_INSERT),
		"commit",
		"rollback",
	}
	if !reflect.DeepEqual(events, expected) {
		t.Errorf("Expected notifications %v but got %v", expected, events)
	}
}

func TestHooksNil(t *testing.T) {
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

	driverConn.RegisterCommitHook(nil)
	driverConn.RegisterRollbackHook(nil)
	driverConn.RegisterUpdateHook(nil)
}
