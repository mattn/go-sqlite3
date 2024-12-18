//go:build cksumvfs
// +build cksumvfs

package sqlite3

import (
	"testing"
	"os"
	"database/sql"
	"io/ioutil"
	"bytes"
	"errors"
)

func TestCksumVfs(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	sql.Register("sqlite3_with_reserved_bytes", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return conn.SetFileControlInt("", SQLITE_FCNTL_RESERVE_BYTES, 8)
		},
	})

	db, err := sql.Open("sqlite3_with_reserved_bytes", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	InitCksumVFS()

	_, err = db.Exec(`create table foo (v string)`)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	stmt, err := db.Prepare("insert into foo(v) values(?)")
	if err != nil {
		t.Fatal("Failed to prepare insert:", err)
	}
	
	for _, v := range []string{"this-is-the-target-string", "foo", "bar", "baz"} {
		_, err = stmt.Exec(v)
		if err != nil {
			t.Fatal("Failed to insert value:", err)
		}
	}

	stmt.Close()
	db.Close()

	// Corrupt the file by replacing one of the column's values
	data, err := ioutil.ReadFile(tempFilename)
	if err != nil {
		t.Fatal("Failed to read database file as bytes:", err)
	}

	newData := bytes.Replace(data, []byte("this-is-the-target-string"), []byte("This-is-the-target-string"), 1)
	if err := ioutil.WriteFile(tempFilename, newData, 0); err != nil {
		t.Fatal("Failed to write database file as new bytes:", err)
	}

	db, err = sql.Open("sqlite3_with_reserved_bytes", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	InitCksumVFS()

	rows, err := db.Query("SELECT * FROM foo")
	if err != nil {
		t.Fatal("Failed to query database:", err)
	}

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal("Failed to scan row:", err)
		}
	}

	if err := rows.Close(); err != nil {
		t.Fatal("Failed to close rows:", err)
	}

	err = rows.Err()

	var sqliteErr Error
	if !errors.As(err, &sqliteErr) {
		t.Fatal("Failed to get close error as SQLite error:", err)
	}

	if sqliteErr.ExtendedCode != ErrIoErrData {
		t.Fatal("Expected extended error of ERR_IO_ERR_DATA, but got:", int(sqliteErr.ExtendedCode))
	}
}
