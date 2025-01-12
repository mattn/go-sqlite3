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
			t.Errorf("Failed to delete session: %v", err)
		}
	}()

	err = session.AttachSession("test_table")
	if err != nil {
		t.Fatalf("Failed to attach session to table: %v", err)
	}

	_, err = conn.ExecContext(ctx, `INSERT INTO test_table (value) VALUES ('test')`)
	if err != nil {
		t.Fatalf("Failed to insert row: %v", err)
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
	} else if !b {
		t.Fatalf("changeset does not contain changes: %v", b)
	}

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

	dest := make([]any, nCol)
	if err := iter.New(dest); err != nil {
		t.Fatalf("Failed to get new row: %v", err)
	}
	if v, ok := dest[0].(int64); !ok {
		t.Fatalf("Expected int64, got %T", dest[0])
	} else if v != 1 {
		t.Fatalf("Expected 1, got %d", v)
	}
	if v, ok := dest[1].(string); !ok {
		t.Fatalf("Expected string, got %T", dest[1])
	} else if v != "test" {
		t.Fatalf("Expected test, got %s", v)
	}

	if b, err := iter.Next(); err != nil {
		t.Fatalf("Failed to get next changeset: %v", err)
	} else if b {
		t.Fatalf("changeset contains more changes: %v", b)
	}

	if err := iter.Finalize(); err != nil {
		t.Fatalf("Failed to finalize changeset iterator: %v", err)
	}
}
