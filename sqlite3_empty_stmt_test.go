// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
	"time"
)

// runBounded runs f in a goroutine and fails the test if it does not return
// within timeout. Guards against the driver returning nil from Next without a
// row, which would otherwise make database/sql loop forever.
func runBounded(t *testing.T, timeout time.Duration, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("operation did not complete within %v (possible infinite loop)", timeout)
	}
}

// TestEmptyQueryConnLevel verifies the conn-level Query("") path: rows reported
// immediately, no rows, no error. This was already handled correctly (see
// SQLiteConn.query) but is kept as a regression guard.
func TestEmptyQueryConnLevel(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var iter int
	var rowsErr error
	runBounded(t, 2*time.Second, func() {
		rows, err := db.Query("")
		if err != nil {
			t.Errorf("Query(\"\") error: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			iter++
		}
		rowsErr = rows.Err()
	})
	if iter != 0 {
		t.Errorf("expected 0 rows, got %d", iter)
	}
	if rowsErr != nil {
		t.Errorf("unexpected rows error: %v", rowsErr)
	}

	// Also verify whitespace-only and comment-only input, which sqlite3_prepare_v2
	// treats the same as empty.
	for _, q := range []string{"   \n\t", "-- just a comment\n", "/* block */;"} {
		var n int
		var rowsErr error
		runBounded(t, 2*time.Second, func() {
			rows, err := db.Query(q)
			if err != nil {
				t.Errorf("Query(%q) error: %v", q, err)
				return
			}
			defer rows.Close()
			for rows.Next() {
				n++
			}
			rowsErr = rows.Err()
		})
		if n != 0 {
			t.Errorf("Query(%q): expected 0 rows, got %d", q, n)
		}
		if rowsErr != nil {
			t.Errorf("Query(%q): unexpected rows error: %v", q, rowsErr)
		}
	}
}

// TestEmptyQueryPrepareThenQuery verifies the stmt-level path that previously
// segfaulted: Prepare("") succeeds (sqlite3_prepare_v2 returns SQLITE_OK with a
// NULL handle) and stmt.Query() must return rows that EOF immediately rather
// than crashing or looping.
func TestEmptyQueryPrepareThenQuery(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("")
	if err != nil {
		t.Fatalf("Prepare(\"\") error: %v", err)
	}
	defer stmt.Close()

	var iter int
	var rowsErr error
	runBounded(t, 2*time.Second, func() {
		rows, err := stmt.Query()
		if err != nil {
			t.Errorf("stmt.Query() error: %v", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			iter++
		}
		rowsErr = rows.Err()
	})
	if iter != 0 {
		t.Errorf("expected 0 rows, got %d", iter)
	}
	if rowsErr != nil {
		t.Errorf("unexpected rows error: %v", rowsErr)
	}
}

// TestEmptyExecPrepareThenExec verifies the stmt-level Exec path that previously
// segfaulted: Prepare("") + stmt.Exec() must succeed and report a zero effect.
func TestEmptyExecPrepareThenExec(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("")
	if err != nil {
		t.Fatalf("Prepare(\"\") error: %v", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		t.Fatalf("stmt.Exec() error: %v", err)
	}
	if id, _ := res.LastInsertId(); id != 0 {
		t.Errorf("expected LastInsertId 0, got %d", id)
	}
	if n, _ := res.RowsAffected(); n != 0 {
		t.Errorf("expected RowsAffected 0, got %d", n)
	}
}

// TestEmptyStatementDriverLevel exercises the driver API directly (bypassing
// database/sql) for the exact scenario described: Prepare("") then Query must
// return rows whose Next reports io.EOF with no columns, not segfault or loop.
func TestEmptyStatementDriverLevel(t *testing.T) {
	c, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	conn := c.(*SQLiteConn)

	stmt, err := conn.Prepare("")
	if err != nil {
		t.Fatalf("Prepare(\"\") error: %v", err)
	}
	defer stmt.Close()

	if n := stmt.NumInput(); n != 0 {
		t.Errorf("NumInput: expected 0, got %d", n)
	}
	// Readonly is only exposed on the concrete type, not driver.Stmt.
	if !stmt.(*SQLiteStmt).Readonly() {
		t.Errorf("Readonly: expected true for empty statement")
	}

	rows, err := stmt.Query(nil)
	if err != nil {
		t.Fatalf("stmt.Query() error: %v", err)
	}
	defer rows.Close()

	ss := rows.(*SQLiteRows)
	if cols := ss.Columns(); len(cols) != 0 {
		t.Errorf("Columns: expected empty, got %v", cols)
	}

	// Next must terminate with io.EOF rather than looping or returning nil.
	var nextErr error
	runBounded(t, 2*time.Second, func() {
		dest := make([]driver.Value, 0)
		nextErr = ss.Next(dest)
	})
	if nextErr != io.EOF {
		t.Errorf("Next: expected io.EOF, got %v", nextErr)
	}
}

// TestEmptyStatementConnExec verifies conn-level Exec("") (no-args fast path and
// general path) both succeed with a zero-effect result.
func TestEmptyStatementConnExec(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, q := range []string{"", "   ", "-- comment\n"} {
		res, err := db.Exec(q)
		if err != nil {
			t.Errorf("Exec(%q) error: %v", q, err)
			continue
		}
		if n, _ := res.RowsAffected(); n != 0 {
			t.Errorf("Exec(%q): expected 0 rows affected, got %d", q, n)
		}
	}
}
