// +build !sqlite_omit_deserialize

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"os"
	"testing"
)

func TestSerialize(t *testing.T) {
	d := SQLiteDriver{}

	srcConn, err := d.Open(":memory:")
	if err != nil {
		t.Fatal("failed to get database connection:", err)
	}
	defer srcConn.Close()
	sqlite3conn := srcConn.(*SQLiteConn)

	_, err = sqlite3conn.Exec(`CREATE TABLE foo (name string)`, nil)
	if err != nil {
		t.Fatal("failed to create table:", err)
	}
	_, err = sqlite3conn.Exec(`INSERT INTO foo(name) VALUES("alice")`, nil)
	if err != nil {
		t.Fatal("failed to insert record:", err)
	}

	// Serialize the database to a file
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	b, err := sqlite3conn.Serialize("")
	if err != nil {
		t.Fatalf("failed to serialize database: %s", err)
	}
	if err := os.WriteFile(tempFilename, b, 0644); err != nil {
		t.Fatalf("failed to write serialized database to disk")
	}

	// Open the SQLite3 file, and test that contents are as expected.
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT * FROM foo`)
	if err != nil {
		t.Fatal("failed to query database:", err)
	}
	defer rows.Close()

	rows.Next()

	var name string
	rows.Scan(&name)
	if exp, got := name, "alice"; exp != got {
		t.Errorf("Expected %s for fetched result, but got %s:", exp, got)
	}
}

func TestDeserialize(t *testing.T) {
	var sqlite3conn *SQLiteConn
	d := SQLiteDriver{}
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	// Create source database on disk.
	conn, err := d.Open(tempFilename)
	if err != nil {
		t.Fatal("failed to open on-disk database:", err)
	}
	defer conn.Close()
	sqlite3conn = conn.(*SQLiteConn)
	_, err = sqlite3conn.Exec(`CREATE TABLE foo (name string)`, nil)
	if err != nil {
		t.Fatal("failed to create table:", err)
	}
	_, err = sqlite3conn.Exec(`INSERT INTO foo(name) VALUES("alice")`, nil)
	if err != nil {
		t.Fatal("failed to insert record:", err)
	}
	conn.Close()

	// Read database file bytes from disk.
	b, err := os.ReadFile(tempFilename)
	if err != nil {
		t.Fatal("failed to read database file on disk", err)
	}

	// Deserialize file contents into memory.
	conn, err = d.Open(":memory:")
	if err != nil {
		t.Fatal("failed to open in-memory database:", err)
	}
	sqlite3conn = conn.(*SQLiteConn)
	defer conn.Close()
	if err := sqlite3conn.Deserialize(b, ""); err != nil {
		t.Fatal("failed to deserialize database", err)
	}

	// Check database contents are as expected.
	rows, err := sqlite3conn.Query(`SELECT * FROM foo`, nil)
	if err != nil {
		t.Fatal("failed to query database:", err)
	}
	defer rows.Close()
	if len(rows.Columns()) != 1 {
		t.Fatal("incorrect number of columns returned:", len(rows.Columns()))
	}
	values := make([]driver.Value, 1)
	rows.Next(values)
	if v, ok := values[0].(string); !ok {
		t.Fatalf("wrong type for value: %T", v)
	} else if v != "alice" {
		t.Fatal("wrong value returned", v)
	}
}
