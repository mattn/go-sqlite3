package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"os"
	"strings"
	"testing"
)

// Register ConnectHook, get *SQLiteRows from driver.Rows,
// test DeclTypes() method.
func TestDeclTypes(t *testing.T) {
	var sqlite3conn *SQLiteConn = nil
	sql.Register("sqlite3_with_hook_for_decltypes",
			&SQLiteDriver{
					ConnectHook: func(conn *SQLiteConn) error {
						sqlite3conn = conn
						return nil
					},
			})

	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3_with_hook_for_decltypes", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
	defer db.Close()

	db.Exec("CREATE TABLE foo (id INTEGER, d DOUBLE, dt DATETIME, b BLOB);")
	db.Exec("INSERT INTO foo VALUES(1, 0.5, '2015-03-05 15:16:17', x'900df00d');")

	if sqlite3conn == nil {
		t.Fatal("Failed to hook into SQLiteConn")
	}
	rows, err := sqlite3conn.Query("SELECT * FROM foo", []driver.Value{})
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}
	defer rows.Close()

	var decltypes []string
	switch rows.(type) {
	case *SQLiteRows: decltypes = rows.(*SQLiteRows).DeclTypes()
	default: t.Fatal("Failed to convert driver.Rows to *SQLiteRows")
	}

	expectedDeclTypes := []string{"INTEGER", "DOUBLE", "DATETIME", "BLOB"}
	if len(decltypes) != len(expectedDeclTypes) {
		t.Errorf("Number of decltypes should be %d, not %d", len(expectedDeclTypes), len(decltypes))
	}
	for i := 0; i < len(expectedDeclTypes); i++ {
		if !strings.EqualFold(decltypes[i], expectedDeclTypes[i]) {
			t.Errorf("decltype of column %d should be %s, not %s", i, expectedDeclTypes[i], decltypes[i])
		}
	}
}
