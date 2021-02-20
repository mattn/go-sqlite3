// +build sqlite_column_metadata

package sqlite3

import "testing"

func TestColumnTableName(t *testing.T) {
	d := SQLiteDriver{}
	conn, err := d.Open(":memory:")
	if err != nil {
		t.Fatal("failed to get database connection:", err)
	}
	defer conn.Close()
	sqlite3conn := conn.(*SQLiteConn)

	_, err = sqlite3conn.Exec(`CREATE TABLE foo (name string)`, nil)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}
	_, err = sqlite3conn.Exec(`CREATE TABLE bar (name string)`, nil)
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	stmt, err := sqlite3conn.Prepare(`SELECT * FROM foo JOIN bar ON foo.name = bar.name`)
	if err != nil {
		t.Fatal(err)
	}

	if exp, got := "foo", stmt.(*SQLiteStmt).ColumnTableName(0); exp != got {
		t.Fatalf("Incorrect table name returned expected: %s, got: %s", exp, got)
	}
	if exp, got := "bar", stmt.(*SQLiteStmt).ColumnTableName(1); exp != got {
		t.Fatalf("Incorrect table name returned expected: %s, got: %s", exp, got)
	}
	if exp, got := "", stmt.(*SQLiteStmt).ColumnTableName(2); exp != got {
		t.Fatalf("Incorrect table name returned expected: %s, got: %s", exp, got)
	}
}
