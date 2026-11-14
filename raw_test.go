package sqlite3_test

import (
	"database/sql"
	"testing"

	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/mattn/go-sqlite3/internal/sqlite3test"
)

func TestRaw(t *testing.T) {
	sql.Register("sqlite3_rawtest", &sqlite3.SQLiteDriver{
		ConnectHook: func(c *sqlite3.SQLiteConn) error {
			return sqlite3test.RegisterFunction(c)
		},
	})

	db, err := sql.Open("sqlite3_rawtest", "...")
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	var result int
	if err := db.QueryRow(`SELECT one()`).Scan(&result); err != nil {
		t.Fatal(err)
	}

	if result != 1 {
		t.Errorf("expected custom one() function to return 1, but got %d", result)
	}
}
