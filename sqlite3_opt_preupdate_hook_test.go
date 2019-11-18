// Copyright (C) 2019 G.J.R. Timmer <gjr.timmer@gmail.com>.
// Copyright (C) 2018 segment.com <friends@segment.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_preupdate_hook

package sqlite3

import (
	"database/sql"
	"testing"
)

type preUpdateHookDataForTest struct {
	databaseName string
	tableName    string
	count        int
	op           int
	oldRow       []interface{}
	newRow       []interface{}
}

func TestPreUpdateHook(t *testing.T) {
	var events []preUpdateHookDataForTest

	sql.Register("sqlite3_PreUpdateHook", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			conn.RegisterPreUpdateHook(func(data SQLitePreUpdateData) {
				eval := -1
				oldRow := []interface{}{eval}
				if data.Op != SQLITE_INSERT {
					err := data.Old(oldRow...)
					if err != nil {
						t.Fatalf("Unexpected error calling SQLitePreUpdateData.Old: %v", err)
					}
				}

				eval2 := -1
				newRow := []interface{}{eval2}
				if data.Op != SQLITE_DELETE {
					err := data.New(newRow...)
					if err != nil {
						t.Fatalf("Unexpected error calling SQLitePreUpdateData.New: %v", err)
					}
				}

				// tests dest bound checks in loop
				var tooSmallRow []interface{}
				if data.Op != SQLITE_INSERT {
					err := data.Old(tooSmallRow...)
					if err != nil {
						t.Fatalf("Unexpected error calling SQLitePreUpdateData.Old: %v", err)
					}
					if len(tooSmallRow) != 0 {
						t.Errorf("Expected tooSmallRow to be empty, got: %v", tooSmallRow)
					}
				}

				events = append(events, preUpdateHookDataForTest{
					databaseName: data.DatabaseName,
					tableName:    data.TableName,
					count:        data.Count(),
					op:           data.Op,
					oldRow:       oldRow,
					newRow:       newRow,
				})
			})
			return nil
		},
	})

	db, err := sql.Open("sqlite3_PreUpdateHook", ":memory:")
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

	if len(events) != 3 {
		t.Errorf("Events should be 3 entries, got: %d", len(events))
	}

	if events[0].op != SQLITE_INSERT {
		t.Errorf("Op isn't as expected: %v", events[0].op)
	}

	if events[1].op != SQLITE_UPDATE {
		t.Errorf("Op isn't as expected: %v", events[1].op)
	}

	if events[1].count != 1 {
		t.Errorf("Expected event row 1 to have 1 column, had: %v", events[1].count)
	}

	newRow_0_0 := events[0].newRow[0].(int64)
	if newRow_0_0 != 9 {
		t.Errorf("Expected event row 0 new column 0 to be == 9, got: %v", newRow_0_0)
	}

	oldRow_1_0 := events[1].oldRow[0].(int64)
	if oldRow_1_0 != 9 {
		t.Errorf("Expected event row 1 old column 0 to be == 9, got: %v", oldRow_1_0)
	}

	newRow_1_0 := events[1].newRow[0].(int64)
	if newRow_1_0 != 99 {
		t.Errorf("Expected event row 1 new column 0 to be == 99, got: %v", newRow_1_0)
	}

	oldRow_2_0 := events[2].oldRow[0].(int64)
	if oldRow_2_0 != 99 {
		t.Errorf("Expected event row 1 new column 0 to be == 99, got: %v", oldRow_2_0)
	}
}
