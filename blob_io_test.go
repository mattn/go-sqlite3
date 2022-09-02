// Copyright (C) 2022 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo
// +build cgo

package sqlite3

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"testing"
)

func TestBlobIO(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:testblobio?mode=memory&cache=shared")
	if err != nil {
		t.Fatal("Fail to open:", err)
	}
	defer db.Close()

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal("Failed to get raw connection:", err)
	}
	defer conn.Close()

	var driverConn *SQLiteConn
	err = conn.Raw(func(conn interface{}) error {
		driverConn = conn.(*SQLiteConn)
		return nil
	})
	if err != nil {
		t.Fatal("Failed to get raw connection:", err)
	}
	defer driverConn.Close()

	// Test data
	expected := []byte("I ❤️ SQLite in \x00\x01\x02…")
	rowid := int64(6581)

	query := `
		CREATE TABLE data (
			value BLOB
		);

		INSERT INTO data (_rowid_, value)
			VALUES (:rowid, :value);
	`

	_, err = db.Exec(query, sql.Named("rowid", rowid), sql.Named("value", expected))
	if err != nil {
		t.Fatal("Failed to execute", err)
	}

	// Open blob
	blob, err := driverConn.Blob("main", "data", "value", rowid, 0)
	if err != nil {
		t.Error("failed", err)
	}
	defer blob.Close()

	// Read blob incrementally
	middle := len(expected) / 2
	first := expected[:middle]
	second := expected[middle:]

	// Read part Ⅰ
	b1 := make([]byte, len(first))
	n1, err := blob.Read(b1)

	if err != nil || n1 != len(b1) {
		t.Errorf("Failed to read %d bytes", n1)
	}

	if bytes.Compare(first, b1) != 0 {
		t.Error("Expected\n", first, "got\n", b1)
	}

	// Read part Ⅱ
	b2 := make([]byte, len(second))
	n2, err := blob.Read(b2)

	if err != nil || n2 != len(b2) {
		t.Errorf("Failed to read %d bytes", n2)
	}

	if bytes.Compare(second, b2) != 0 {
		t.Error("Expected\n", second, "got\n", b2)
	}

	// EOF
	b3 := make([]byte, 10)
	n3, err := blob.Read(b3)

	if err != io.EOF || n3 != 0 {
		t.Error("Expected EOF", err)
	}
}
