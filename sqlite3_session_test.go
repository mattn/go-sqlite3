package sqlite3

import (
	"context"
	"database/sql"
	"testing"
)

func Test_EmptyChangeset(t *testing.T) {
	// Open a new in-memory SQLite database
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

	changeset, err := session.Changeset()
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
