//go:build sqlite_vtable

package sqlite3

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
)

var leakCheckDriverSeq int32

func TestVtabCursorHandleRelease(t *testing.T) {
	// Use a unique driver name so repeated runs (e.g. -count=2) do not
	// panic on duplicate registration.
	driverName := fmt.Sprintf("sqlite3_HandleLeakCheck_%d", atomic.AddInt32(&leakCheckDriverSeq, 1))
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return conn.CreateModule("test", &testModule{t: t, intarray: []int{1, 2, 3}})
		},
	})
	db, err := sql.Open(driverName, ":memory:")
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
