// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build !libsqlite3

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestLimit(t *testing.T) {
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

	if rv := driverConn.SetLimit(SQLITE_LIMIT_TRIGGER_DEPTH, 5); rv != 1000 && rv != -1 {
		t.Fatalf("Unable to set limit; %d", rv)
	}

	if limit := driverConn.GetLimit(SQLITE_LIMIT_TRIGGER_DEPTH); limit != 5 {
		t.Fatal("Limit was not set correctly")
	}
}
