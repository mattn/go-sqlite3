//go:build !libsqlite3 || sqlite_serialize
// +build !libsqlite3 sqlite_serialize

package sqlite3

import (
	"database/sql"
	"os"
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	// The driver's connection will be needed in order to serialization and deserialization
	driverName := "TestSerializeDeserialize"
	driverConns := []*SQLiteConn{}
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			driverConns = append(driverConns, conn)
			return nil
		},
	})

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

	// Check the driver connections.
	if len(driverConns) != 2 {
		t.Fatalf("Expected 2 driver connections, but found %v.", len(driverConns))
	}
	srcDbDriverConn := driverConns[0]
	if srcDbDriverConn == nil {
		t.Fatal("The source database driver connection is nil.")
	}
	destDbDriverConn := driverConns[1]
	if destDbDriverConn == nil {
		t.Fatal("The destination database driver connection is nil.")
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

	// Serialize source database
	b, err := srcDbDriverConn.Serialize("")
	if err != nil {
		t.Fatalf("Failed to serialize source database: %s", err)
	}

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
	if err := destDbDriverConn.Deserialize(b, ""); err != nil {
		t.Fatal("Failed to deserialize to destination database", err)
	}

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
