//go:build !libsqlite3 || sqlite_serialize
// +build !libsqlite3 sqlite_serialize

package sqlite3

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	// Connect to the source database.
	srcTempFilename := TempFilename(t)
	defer os.Remove(srcTempFilename)
	srcDb, err := sql.Open(driverName, srcTempFilename)
	if err != nil {
		t.Fatal("Failed to open the source database:", err)
	}
	defer srcDb.Close()
	err = srcDb.Ping()
	if err != nil {
		t.Fatal("Failed to connect to the source database:", err)
	}

	// Connect to the destination database.
	destTempFilename := TempFilename(t)
	defer os.Remove(destTempFilename)
	destDb, err := sql.Open(driverName, destTempFilename)
	if err != nil {
		t.Fatal("Failed to open the destination database:", err)
	}
	defer destDb.Close()
	err = destDb.Ping()
	if err != nil {
		t.Fatal("Failed to connect to the destination database:", err)
	}

	// Write data to source database.
	_, err = srcDb.Exec(`CREATE TABLE foo (name string)`)
	if err != nil {
		t.Fatal("Failed to create table in source database:", err)
	}
	_, err = srcDb.Exec(`INSERT INTO foo(name) VALUES("alice")`)
	if err != nil {
		t.Fatal("Failed to insert data into source database", err)
	}

	// Serialize the source database
	srcConn, err := srcDb.Conn(context.Background())
	if err != nil {
		t.Fatal("Failed to get connection to source database:", err)
	}
	defer srcConn.Close()

	var serialized []byte
	if err := srcConn.Raw(func(raw any) error {
		var err error
		serialized, err = raw.(*SQLiteConn).Serialize("")
		return err
	}); err != nil {
		t.Fatal("Failed to serialize source database:", err)
	}
	srcConn.Close()

	// Confirm that the destination database is initially empty.
	var destTableCount int
	err = destDb.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table'").Scan(&destTableCount)
	if err != nil {
		t.Fatal("Failed to check the destination table count:", err)
	}
	if destTableCount != 0 {
		t.Fatalf("The destination database is not empty; %v table(s) found.", destTableCount)
	}

	// Deserialize to destination database
	destConn, err := destDb.Conn(context.Background())
	if err != nil {
		t.Fatal("Failed to get connection to destination database:", err)
	}
	defer destConn.Close()

	if err := destConn.Raw(func(raw any) error {
		return raw.(*SQLiteConn).Deserialize(serialized, "")
	}); err != nil {
		t.Fatal("Failed to deserialize source database:", err)
	}
	destConn.Close()

	// Confirm that destination database has been loaded correctly.
	var destRowCount int
	err = destDb.QueryRow(`SELECT COUNT(*) FROM foo`).Scan(&destRowCount)
	if err != nil {
		t.Fatal("Failed to count rows in destination database table", err)
	}
	if destRowCount != 1 {
		t.Fatalf("Destination table does not have the expected records")
	}
}
