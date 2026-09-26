//go:build sqlite_session
// +build sqlite_session

package sqlite3

import (
	"context"
	"database/sql/driver"
	"io"
	"testing"
)

func TestSessionChangesetApply(t *testing.T) {
	// Use the SQLiteDriver directly to easily get SQLiteConn
	var d SQLiteDriver // Assuming SQLiteDriver implements driver.Driver

	// Use ":memory:" for simplicity
	db1, err := d.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open db1: %v", err)
	}
	// Assert the driver.Conn to the specific type *SQLiteConn which has session methods
	// Ensure SQLiteConn implements driver.ConnBeginTx and driver.ExecerContext etc. if needed elsewhere
	conn1, ok := db1.(driver.ExecerContext) // Check needed interfaces first
	if !ok {
		db1.Close()
		t.Fatalf("db1 connection does not implement driver.ExecerContext")
	}
	sqliteConn1, ok := db1.(*SQLiteConn)
	if !ok {
		db1.Close()
		t.Fatalf("db1 connection is not *SQLiteConn")
	}
	defer sqliteConn1.Close() // Use the concrete type for Close if it has specific logic

	db2, err := d.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to open db2: %v", err)
	}
	conn2, ok := db2.(driver.QueryerContext) // Need QueryerContext for QueryContext
	if !ok {
		db2.Close()
		t.Fatalf("db2 connection does not implement driver.QueryerContext")
	}
	sqliteConn2, ok := db2.(*SQLiteConn) // Also need *SQLiteConn for ChangesetApply
	if !ok {
		db2.Close()
		t.Fatalf("db2 connection is not *SQLiteConn")
	}
	defer sqliteConn2.Close()

	// --- Setup Database 1 ---
	tableName := "users"
	_, err = conn1.ExecContext(context.Background(), `CREATE TABLE `+tableName+` (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`, nil)
	if err != nil {
		t.Fatalf("db1: Failed to create table '%s': %v", tableName, err)
	}

	// --- Create Session and Make Changes ---
	session, err := sqliteConn1.SessionCreate("main")
	if err != nil {
		t.Fatalf("db1: Failed to create session: %v", err)
	}
	// defer session.Delete() // If applicable

	err = session.Attach(tableName)
	if err != nil {
		t.Fatalf("db1: Failed to attach session to table '%s': %v", tableName, err)
	}

	insertName := "Alice"
	// Use []driver.NamedValue for parameters with ExecContext/QueryContext
	args := []driver.NamedValue{{Name: "", Ordinal: 1, Value: insertName}}
	_, err = conn1.ExecContext(context.Background(), `INSERT INTO `+tableName+` (name) VALUES (?)`, args)
	if err != nil {
		t.Fatalf("db1: Failed to insert data into '%s': %v", tableName, err)
	}

	// --- Generate Changeset ---
	changeset, err := session.Changeset()
	if err != nil {
		t.Fatalf("db1: Failed to get changeset: %v", err)
	}
	if session.IsEmpty() {
		t.Fatal("db1: Session is unexpectedly empty after insert")
	}

	// --- Setup Database 2 ---
	conn2Execer, ok := db2.(driver.ExecerContext) // Need ExecerContext for CREATE TABLE
	if !ok {
		t.Fatalf("db2 connection does not implement driver.ExecerContext")
	}
	_, err = conn2Execer.ExecContext(context.Background(), `CREATE TABLE `+tableName+` (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`, nil)
	if err != nil {
		t.Fatalf("db2: Failed to create table '%s': %v", tableName, err)
	}

	// --- Apply Changeset to Database 2 ---
	// Use the *SQLiteConn for driver-specific methods
	rc := sqliteConn2.ChangesetApply(changeset, nil, nil)
	if rc != 0 {
		t.Fatalf("db2: ChangesetApply failed with code %d", rc)
	}

	// --- Verification ---
	// Check if the row was inserted into the second database via the changeset.
	// Query using the driver.QueryerContext interface
	queryArgs := []driver.NamedValue{{Name: "", Ordinal: 1, Value: insertName}}
	rows, err := conn2.QueryContext(context.Background(), `SELECT id, name FROM `+tableName+` WHERE name = ?`, queryArgs)
	if err != nil {
		t.Fatalf("db2: Failed to query table '%s' after applying changeset: %v", tableName, err)
	}
	// Defer close immediately after successful query
	// Use t.Errorf for close error to avoid masking the main test failure.
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			t.Errorf("db2: Error closing rows: %v", cerr)
		}
	}()

	columns := rows.Columns()
	if len(columns) != 2 { // Expecting id and name
		t.Fatalf("db2: Expected 2 columns (id, name), got %d: %v", len(columns), columns)
	}
	// Find column indices (more robust than assuming order)
	idColIdx := -1
	nameColIdx := -1
	for i, colName := range columns {
		switch colName {
		case "id":
			idColIdx = i
		case "name":
			nameColIdx = i
		}
	}
	if idColIdx == -1 || nameColIdx == -1 {
		t.Fatalf("db2: Could not find expected columns 'id' and 'name' in results: %v", columns)
	}

	values := make([]driver.Value, len(columns))
	rowCount := 0
	var nameFound string
	var idFound int64 // Assuming id is INTEGER -> int64

	// Use the iteration pattern provided
	for {
		err = rows.Next(values) // Fetch the next row's data into the values slice
		if err == io.EOF {
			break // End of rows
		}
		if err != nil {
			// Close was deferred, just fail the test
			t.Fatalf("db2: Error iterating rows: %v", err)
		}

		rowCount++
		if rowCount > 1 {
			// Optionally log the extra row data before failing
			t.Logf("db2: Extra row data - %v", values)
			t.Fatalf("db2: Expected only 1 row, but found more.")
		}

		// Extract values using the found indices
		idValue, idOk := values[idColIdx].(int64)
		nameValue, nameOk := values[nameColIdx].(string)

		if !idOk {
			// Log the actual type for easier debugging
			t.Fatalf("db2: Unexpected data type for 'id' column. Expected int64, Got: %T (%v)", values[idColIdx], values[idColIdx])
		}
		if !nameOk {
			t.Fatalf("db2: Unexpected data type for 'name' column. Expected string, Got: %T (%v)", values[nameColIdx], values[nameColIdx])
		}

		idFound = idValue
		nameFound = nameValue

		// You can log the row here if needed for debugging
		// t.Logf("db2: Processing row: id=%v, name=%v", idFound, nameFound)
	}

	// Check results *after* the loop
	if rowCount == 0 {
		t.Fatal("db2: No rows found after applying changeset, expected 1 row.")
	}
	// rowCount == 1 is implicitly checked by the loop logic preventing > 1

	if nameFound != insertName {
		t.Fatalf("db2: Expected name '%s', but found '%s' (id: %d)", insertName, nameFound, idFound)
	}

	// rows.Err() is typically checked after a sql.Rows loop, but the driver.Rows
	// Next() method should return any errors directly, making a final rows.Err() check less critical here.
	// The loop already breaks/fails on non-EOF errors from rows.Next().

	t.Log("Changeset successfully generated, applied, and verified using driver interfaces.")
}
