//go:build sqlite_vtable

package sqlite3

import (
	"database/sql"
	"testing"
)

func TestVtabCursorHandleRelease(t *testing.T) {
	sql.Register("sqlite3_HandleLeakCheck", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return conn.CreateModule("test", &testModule{t: t, intarray: []int{1, 2, 3}})
		},
	})
	db, err := sql.Open("sqlite3_HandleLeakCheck", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE VIRTUAL TABLE vtab USING test('1', 2, three)"); err != nil {
		t.Fatal(err)
	}
	var before, after int
	for i := 0; i < 50; i++ {
		var n int
		if err := db.QueryRow("SELECT count(*) FROM vtab").Scan(&n); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			before = len(loadHandleVals())
		}
	}
	after = len(loadHandleVals())
	if after > before {
		t.Fatalf("handle map grew from %d to %d over repeated cursor open/close", before, after)
	}
}
