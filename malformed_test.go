package sqlite3

import (
	"context"
	"database/sql"
	"testing"
)

func Test_TableCreationInMemoryLoadRaw(t *testing.T) {
	rwDSN := "file:/mydb?mode=rw&vfs=memdb&_txlock=immediate&_fk=false"
	roDSN := "file:/mydb?mode=ro&vfs=memdb&_txlock=deferred&_fk=false"

	rwDB, err := sql.Open("sqlite3", rwDSN)
	if err != nil {
		t.Fatalf("failed to open rw database: %s", err)
	}
	roDB, err := sql.Open("sqlite3", roDSN)
	if err != nil {
		t.Fatalf("failed to open ro database: %s", err)
	}

	rwConn, err := rwDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("failed to create rw connection: %s", err)
	}
	defer rwConn.Close()
	roConn, err := roDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("failed to create ro connection: %s", err)
	}
	defer roConn.Close()

	_, err = rwConn.ExecContext(context.Background(), `CREATE TABLE logs (entry TEXT)`, nil)
	if err != nil {
		t.Fatalf("failed to create table: %s", err)
	}

	go func() {
		for {
			var err error
			_, err = rwConn.ExecContext(context.Background(), `INSERT INTO logs(entry) VALUES("hello")`, nil)
			if err != nil {
				return
			}
		}
	}()

	n := 0
	for {
		n++
		rows, err := roConn.QueryContext(context.Background(), `SELECT COUNT(*) FROM logs`, nil)
		if err != nil {
			t.Fatalf("failed to query for count: %s", err)
		}

		for rows.Next() {
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("rows had error after Next() (%d loops): %s", n, err)

		}
		rows.Close()
	}
}

