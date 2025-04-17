// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo
// +build cgo

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

func TestDbpage(t *testing.T) {
	sourceFilename := TempFilename(t)
	defer os.Remove(sourceFilename)

	destFilename := TempFilename(t)
	defer os.Remove(destFilename)

	db, err := sql.Open("sqlite3", sourceFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatal("Failed to Exec PRAGMA journal_mode:", err)
	} else if _, err := db.Exec("CREATE TABLE foo(data TEXT)"); err != nil {
		t.Fatal("Failed to create table:", err)
	} else if _, err := db.Exec("INSERT INTO foo(data) VALUES(?)", "hello sqlite_dbpage"); err != nil {
		t.Fatal("Failed to create table:", err)
	}

	rows, err := db.Query("SELECT data FROM sqlite_dbpage")
	if err != nil && err.Error() == "no such table: sqlite_dbpage" {
		t.Skip("sqlite_dbpage not supported")
	} else if err != nil {
		t.Fatal("Unable to query sqlite_dbpage table:", err)
	}
	defer rows.Close()

	destFile, err := os.OpenFile(destFilename, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		t.Fatal("Unable to open file for writing:", err)
	}
	defer destFile.Close()

	for rows.Next() {
		var page []byte
		if err := rows.Scan(&page); err != nil {
			t.Fatal("Unable to scan results:", err)
		}

		if _, err := destFile.Write(page); err != nil {
			t.Fatal("Unable to write page to file:", err)
		}
	}
	if err := rows.Close(); err != nil {
		t.Fatal("Unable to close rows:", err)
	} else if err := db.Close(); err != nil {
		t.Fatal("Unable to close database:", err)
	} else if err := destFile.Close(); err != nil {
		t.Fatal("Unable to close file:", err)
	}

	db, err = sql.Open("sqlite3", destFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	var result string
	if err = db.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		t.Fatal("Failed to query PRAGMA integrity_check:", err)
	} else if result != "ok" {
		t.Fatal("Copied database integrity check failed:", result)
	}

	var hello string
	if err = db.QueryRow("SELECT data FROM foo").Scan(&hello); err != nil {
		t.Fatal("Failed to query data:", err)
	} else if hello != "hello sqlite_dbpage" {
		t.Fatal("Unable to find expected data:", hello)
	}
}
