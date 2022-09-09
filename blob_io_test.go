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

// Verify interface implementations
var _ io.Reader = &SQLiteBlob{}
var _ io.Writer = &SQLiteBlob{}
var _ io.Seeker = &SQLiteBlob{}
var _ io.Closer = &SQLiteBlob{}

type driverConnCallback func(*testing.T, *SQLiteConn)

func blobTestData(t *testing.T, dbname string, rowid int64, blob []byte, c driverConnCallback) error {
	db, err := sql.Open("sqlite3", "file:"+dbname+"?mode=memory&cache=shared")
	if err != nil {
		return err
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()

	var driverConn *SQLiteConn
	err = conn.Raw(func(conn interface{}) error {
		driverConn = conn.(*SQLiteConn)
		return nil
	})
	if err != nil {
		return err
	}
	defer driverConn.Close()

	query := `
		CREATE TABLE data (
			value BLOB
		);

		INSERT INTO data (_rowid_, value)
			VALUES (:rowid, :value);
	`

	_, err = db.Exec(query, sql.Named("rowid", rowid), sql.Named("value", blob))
	if err != nil {
		return err
	}

	c(t, driverConn)

	return nil
}

func TestBlobRead(t *testing.T) {
	rowid := int64(6581)
	expected := []byte("I ❤️ SQLite in \x00\x01\x02…")

	err := blobTestData(t, "testblobread", rowid, expected, func(t *testing.T, driverConn *SQLiteConn) {

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
	})

	if err != nil {
		t.Fatal("Failed to get raw connection:", err)
	}
}

func TestBlobWrite(t *testing.T) {
	rowid := int64(8580)
	expected := []byte{
		// Random data from /dev/urandom
		0xe5, 0x48, 0x94, 0xad, 0xa6, 0x7c, 0x81, 0xa2, 0x70, 0x07, 0x79, 0x60,
		0x33, 0xbc, 0x64, 0x33, 0x8f, 0x48, 0x43, 0xa6, 0x33, 0x5c, 0x08, 0x32,
	}

	// Allocate a zero blob
	data := make([]byte, len(expected))
	err := blobTestData(t, "testblobwrite", rowid, data, func(t *testing.T, driverConn *SQLiteConn) {

		// Open blob for read/write
		blob, err := driverConn.Blob("main", "data", "value", rowid, 1)
		if err != nil {
			t.Error("failed", err)
		}
		defer blob.Close()

		// Write blob incrementally
		middle := len(expected) / 2
		first := expected[:middle]
		second := expected[middle:]

		// Write part Ⅰ
		n1, err := blob.Write(first)

		if err != nil || n1 != len(first) {
			t.Errorf("Failed to write %d bytes", n1)
		}

		// Write part Ⅱ
		n2, err := blob.Write(second)

		if err != nil || n2 != len(second) {
			t.Errorf("Failed to write %d bytes", n2)
		}

		// EOF
		b3 := make([]byte, 10)
		n3, err := blob.Write(b3)

		if err != io.EOF || n3 != 0 {
			t.Error("Expected EOF", err)
		}

		// Verify written data
		_, err = blob.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatal("Failed to seek:", err)
		}

		b4 := make([]byte, len(expected))
		n4, err := blob.Read(b4)

		if err != nil || n4 != len(b4) {
			t.Errorf("Failed to read %d bytes", n4)
		}

		if bytes.Compare(expected, b4) != 0 {
			t.Error("Expected\n", expected, "got\n", b4)
		}

	})
	if err != nil {
		t.Fatal("Failed to get raw connection:", err)
	}
}

func TestBlobSeek(t *testing.T) {
	rowid := int64(6510)
	data := make([]byte, 1000)

	err := blobTestData(t, "testblobseek", rowid, data, func(t *testing.T, driverConn *SQLiteConn) {

		// Open blob
		blob, err := driverConn.Blob("main", "data", "value", rowid, 0)
		if err != nil {
			t.Error("failed", err)
		}
		defer blob.Close()

		// Test data
		begin := int64(0)
		middle := int64(len(data) / 2)
		end := int64(len(data) - 1)
		eof := int64(len(data))

		tests := []struct {
			offset   int64
			whence   int
			expected int64
		}{
			{offset: begin, whence: io.SeekStart, expected: begin},
			{offset: middle, whence: io.SeekStart, expected: middle},
			{offset: end, whence: io.SeekStart, expected: end},
			{offset: eof, whence: io.SeekStart, expected: eof},

			{offset: -1, whence: io.SeekCurrent, expected: middle - 1},
			{offset: 0, whence: io.SeekCurrent, expected: middle},
			{offset: 1, whence: io.SeekCurrent, expected: middle + 1},
			{offset: -middle, whence: io.SeekCurrent, expected: begin},

			{offset: -2, whence: io.SeekEnd, expected: end - 1},
			{offset: -1, whence: io.SeekEnd, expected: end},
			{offset: 0, whence: io.SeekEnd, expected: eof},
			{offset: 1, whence: io.SeekEnd, expected: eof + 1},
			{offset: -eof, whence: io.SeekEnd, expected: begin},
		}

		for _, tc := range tests {
			// Start in the middle
			_, err := blob.Seek(middle, io.SeekStart)
			if err != nil {
				t.Fatal("Failed to seek:", err)
			}

			// Test
			got, err := blob.Seek(tc.offset, tc.whence)
			if err != nil {
				t.Fatal("Failed to seek:", err)
			}

			if tc.expected != got {
				t.Error("For", tc, "expected", tc.expected, "got", got)
			}
		}

	})

	if err != nil {
		t.Fatal("Failed to get raw connection:", err)
	}
}
