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
	"os"
	"testing"
)

func TestPreparedStatementConn(t *testing.T) {
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
	if _, err := db.Exec("CREATE TABLE t (count INT);"); err != nil {
		t.Fatal(err)
	}

	ins, err := driverConn.Prepare("INSERT INTO t (count) VALUES (?)")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}

	qry, err := driverConn.Prepare("SELECT * FROM t WHERE count = ?")
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	for n := 1; n <= 3; n++ {
		if _, err := ins.(*SQLiteStmt).Exec([]driver.Value{n}); err != nil {
			t.Fatalf("insert(%d) = %v", n, err)
		}

		if _, err := qry.(*SQLiteStmt).Query([]driver.Value{n}); err != nil {
			t.Fatalf("query(%d) = %v", n, err)
		}
	}
}
