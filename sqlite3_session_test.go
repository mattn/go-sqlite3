package sqlite3

import (
	"context"
	"database/sql"
	"testing"
)

func Test_EmptyChangeset(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open", err)
	}
	err = db.Ping()
	if err != nil {
		t.Fatal("Failed to ping", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE test_table (id INTEGER PRIMARY KEY, value TEXT);`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal("Failed to get connection to source database:", err)
	}
	defer conn.Close()

	var session *Session
	if err := conn.Raw(func(raw any) error {
		var err error
		session, err = raw.(*SQLiteConn).CreateSession("main")
		return err
	}); err != nil {
		t.Fatal("Failed to serialize source database:", err)
	}
	defer func() {
		if err := session.DeleteSession(); err != nil {
			t.Errorf("Failed to delete session: %v", err)
		}
	}()

	err = session.AttachSession("test_table")
	if err != nil {
		t.Fatalf("Failed to attach session to table: %v", err)
	}

	changeset, err := NewChangeset(session)
	if err != nil {
		t.Fatalf("Failed to generate changeset: %v", err)
	}

	iter, err := NewChangesetIterator(changeset)
	if err != nil {
		t.Fatalf("Failed to create changeset iterator: %v", err)
	}
	if b, err := iter.Next(); err != nil {
		t.Fatalf("Failed to get next changeset: %v", err)
	} else if b {
		t.Fatalf("changeset contains changes: %v", b)
	}

	if err := iter.Finalize(); err != nil {
		t.Fatalf("Failed to finalize changeset iterator: %v", err)
	}
}

func Test_Changeset_OneRow(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open", err)
	}
	err = db.Ping()
	if err != nil {
		t.Fatal("Failed to ping", err)
	}
	defer db.Close()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal("Failed to get connection to database:", err)
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, `CREATE TABLE test_table (id INTEGER PRIMARY KEY, value TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	var session *Session
	if err := conn.Raw(func(raw any) error {
		var err error
		session, err = raw.(*SQLiteConn).CreateSession("main")
		return err
	}); err != nil {
		t.Fatal("Failed to create session:", err)
	}
	defer func() {
		if err := session.DeleteSession(); err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}
	}()

	// Attach to the table, insert a row, and capture the inserted row in a changeset.
	err = session.AttachSession("test_table")
	if err != nil {
		t.Fatalf("Failed to attach session to table: %v", err)
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO test_table (value) VALUES ('fiona')`)
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}
	changeset, err := NewChangeset(session)
	if err != nil {
		t.Fatalf("Failed to generate changeset: %v", err)
	}

	// Prepare to iterate over the changeset.
	iter, err := NewChangesetIterator(changeset)
	if err != nil {
		t.Fatalf("Failed to create changeset iterator: %v", err)
	}
	if b, err := iter.Next(); err != nil {
		t.Fatalf("Failed to get next changeset: %v", err)
	} else if !b {
		t.Fatalf("changeset does not contain changes: %v", b)
	}

	// Check table, number of columns changed, the the operation type.
	tblName, nCol, op, indirect, err := iter.Op()
	if err != nil {
		t.Fatalf("Failed to get changeset operation: %v", err)
	}
	if tblName != "test_table" {
		t.Fatalf("Expected table name 'test_table', got '%s'", tblName)
	}
	if nCol != 2 {
		t.Fatalf("Expected 2 columns, got %d", nCol)
	}
	if op != SQLITE_INSERT {
		t.Fatalf("Expected operation 1, got %d", op)
	}
	if indirect {
		t.Fatalf("Expected indirect false, got true")
	}

	// Now, get the new data.
	dest := make([]any, nCol)
	if err := iter.New(dest); err != nil {
		t.Fatalf("Failed to get new row: %v", err)
	}
	if v, ok := dest[0].(int64); !ok {
		t.Fatalf("Expected int64, got %T", dest[0])
	} else if exp, got := v, int64(1); exp != got {
		t.Fatalf("Expected %d, got %d", exp, got)
	}
	if v, ok := dest[1].(string); !ok {
		t.Fatalf("Expected string, got %T", dest[1])
	} else if exp, got := v, "fiona"; exp != got {
		t.Fatalf("Expected %s, %s", exp, got)
	}

	// We only inserted one row, so there should be no more changes.
	if b, err := iter.Next(); err != nil {
		t.Fatalf("Failed to get next changeset: %v", err)
	} else if b {
		t.Fatalf("changeset contains more changes: %v", b)
	}

	if err := iter.Finalize(); err != nil {
		t.Fatalf("Failed to finalize changeset iterator: %v", err)
	}
}

func Test_Changeset_Multi(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open", err)
	}
	err = db.Ping()
	if err != nil {
		t.Fatal("Failed to ping", err)
	}
	defer db.Close()

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal("Failed to get connection to database:", err)
	}
	defer conn.Close()

	// Create tables, and attach a session.
	mustExecute(conn, `CREATE TABLE table1 (id INTEGER PRIMARY KEY, name TEXT, gpa FLOAT, alive boolean)`)
	mustExecute(conn, `CREATE TABLE table2 (id INTEGER PRIMARY KEY, age INTEGER)`)
	mustExecute(conn, `CREATE TABLE table3 (id INTEGER PRIMARY KEY, company TEXT)`)
	var session *Session
	if err := conn.Raw(func(raw any) error {
		var err error
		session, err = raw.(*SQLiteConn).CreateSession("main")
		return err
	}); err != nil {
		t.Fatal("Failed to create session:", err)
	}
	defer func() {
		if err := session.DeleteSession(); err != nil {
			t.Fatalf("Failed to delete session: %v", err)
		}
	}()
	err = session.AttachSession("")
	if err != nil {
		t.Fatalf("Failed to attach session to table: %v", err)
	}

	// Make a bunch of changes.
	mustExecute(conn, `INSERT INTO table1 (name, gpa, alive) VALUES ('fiona', 3.5, 1)`)
	mustExecute(conn, `INSERT INTO table1 (name, gpa, alive) VALUES ('declan', 2.5, 0)`)
	mustExecute(conn, `INSERT INTO table2 (age) VALUES (20)`)
	mustExecute(conn, `INSERT INTO table2 (age) VALUES (30)`)
	mustExecute(conn, `INSERT INTO table3 (company) VALUES ('foo')`)
	mustExecute(conn, `UPDATE table3 SET company = 'bar' WHERE id = 1`)

	// Prepare to iterate over the changes.
	changeset, err := NewChangeset(session)
	if err != nil {
		t.Fatalf("Failed to generate changeset: %v", err)
	}
	iter, err := NewChangesetIterator(changeset)
	if err != nil {
		t.Fatalf("Failed to create changeset iterator: %v", err)
	}

	tt := []struct {
		table string
		op    int
		old   []any
		new   []any
	}{
		{
			table: "table1",
			op:    SQLITE_INSERT,
			new:   []any{int64(1), "fiona", 3.5, int64(1)},
		},
		{
			table: "table1",
			op:    SQLITE_INSERT,
			new:   []any{int64(2), "declan", 2.5, int64(0)},
		},
		{
			table: "table2",
			op:    SQLITE_INSERT,
			new:   []any{int64(1), int64(20)},
		},
		{
			table: "table2",
			op:    SQLITE_INSERT,
			new:   []any{int64(2), int64(30)},
		},
		{
			table: "table3",
			op:    SQLITE_INSERT,
			new:   []any{int64(1), "bar"},
		},
	}

	for i, v := range tt {
		if b, err := iter.Next(); err != nil {
			t.Fatalf("Failed to get next changeset: %v", err)
		} else if !b {
			t.Fatalf("changeset does not contain changes: %v", b)
		}

		tblName, nCol, op, _, err := iter.Op()
		if err != nil {
			t.Fatalf("Failed to get changeset operation: %v", err)
		}
		if exp, got := v.table, tblName; exp != got {
			t.Fatalf("Expected table name '%s', got '%s'", exp, got)
		}
		// if exp, got := len(v.new), nCol; exp != got {
		// 	t.Fatalf("Expected %d columns, got %d", exp, got)
		// }
		if exp, got := v.op, op; exp != got {
			t.Fatalf("Expected operation %d, got %d", exp, got)
		}

		if v.old != nil {
			dest := make([]any, nCol)
			if err := iter.Old(dest); err != nil {
				t.Fatalf("Failed to get old row: %v", err)
			}
			for j, d := range v.old {
				if exp, got := d, dest[j]; exp != got {
					t.Fatalf("Test %d, expected %v (%T) for dest[%d], got %v (%T)", i, exp, exp, j, got, got)
				}
			}
		}

		if v.new != nil {
			dest := make([]any, nCol)
			if err := iter.New(dest); err != nil {
				t.Fatalf("Failed to get new row: %v", err)
			}
			for j, d := range v.new {
				if exp, got := d, dest[j]; exp != got {
					t.Fatalf("Test %d, expected %v (%T) for new dest[%d], got %v (%T)", i, exp, exp, j, got, got)
				}
			}
		}
	}

	if b, err := iter.Next(); err != nil {
		t.Fatalf("Failed to get next changeset: %v", err)
	} else if b {
		t.Fatalf("changeset contains changes: %v", b)
	}

	if err := iter.Finalize(); err != nil {
		t.Fatalf("Failed to finalize changeset iterator: %v", err)
	}
}

func mustExecute(conn *sql.Conn, stmt string) {
	_, err := conn.ExecContext(context.Background(), stmt)
	if err != nil {
		panic(err)
	}
}
