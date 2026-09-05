// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo

package sqlite3

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"testing"
)

// TestStmtCacheLRUEviction verifies that when the prepared-statement cache is
// full, the least-recently-used entry is evicted to make room for a new one.
// Without eviction, the first N queries to enter the cache would squat on
// every slot forever and any subsequently-prepared query (even a hot one)
// would never benefit from caching.
func TestStmtCacheLRUEviction(t *testing.T) {
	d := SQLiteDriver{}
	conn, err := d.Open(":memory:?_stmt_cache_size=2")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := conn.(*SQLiteConn)
	ctx := context.Background()

	prepareAndClose := func(q string) {
		t.Helper()
		stmt, err := c.prepareWithCache(ctx, q)
		if err != nil {
			t.Fatalf("prepareWithCache(%q): %v", q, err)
		}
		if err := stmt.Close(); err != nil {
			t.Fatalf("Close(%q): %v", q, err)
		}
	}

	q1 := "SELECT 1"
	q2 := "SELECT 2"
	q3 := "SELECT 3"

	// Fill the cache with q1 and q2.
	prepareAndClose(q1)
	prepareAndClose(q2)
	if got, want := len(c.stmtCache), 2; got != want {
		t.Fatalf("after filling: len(stmtCache) = %d, want %d", got, want)
	}
	if cacheCount(c, q1) != 1 || cacheCount(c, q2) != 1 {
		t.Fatalf("after filling: expected q1 and q2 cached, got %#v", cacheKeys(c))
	}

	// Insert q3. q1 is the oldest entry and should be evicted.
	prepareAndClose(q3)
	if got, want := len(c.stmtCache), 2; got != want {
		t.Fatalf("after q3: len(stmtCache) = %d, want %d", got, want)
	}
	if cacheCount(c, q1) != 0 {
		t.Fatalf("after q3: q1 should have been evicted, cache=%#v", cacheKeys(c))
	}
	if cacheCount(c, q2) != 1 || cacheCount(c, q3) != 1 {
		t.Fatalf("after q3: expected q2 and q3 cached, got %#v", cacheKeys(c))
	}

	// Touching q2 should make q3 the oldest (the entry at index 0).
	prepareAndClose(q2)
	if len(c.stmtCache) == 0 || c.stmtCache[0].cacheKey != q3 {
		var head string
		if len(c.stmtCache) > 0 {
			head = c.stmtCache[0].cacheKey
		}
		t.Fatalf("after touching q2: expected q3 at stmtCache[0] (LRU), got %q", head)
	}

	// Insert q1 again. Now q3 should be evicted (q2 is newer).
	prepareAndClose(q1)
	if cacheCount(c, q3) != 0 {
		t.Fatalf("after reinserting q1: q3 should have been evicted, cache=%#v", cacheKeys(c))
	}
	if cacheCount(c, q1) != 1 || cacheCount(c, q2) != 1 {
		t.Fatalf("after reinserting q1: expected q1 and q2 cached, got %#v", cacheKeys(c))
	}
	if got, want := len(c.stmtCache), 2; got != want {
		t.Fatalf("after reinserting q1: len(stmtCache) = %d, want %d", got, want)
	}

	// Sanity-check: no dangling entries past len(stmtCache).
	tail := c.stmtCache[:cap(c.stmtCache)]
	for i := len(c.stmtCache); i < len(tail); i++ {
		if tail[i] != nil {
			t.Fatalf("stmtCache tail slot %d = %p, expected nil", i, tail[i])
		}
	}
}

// TestStmtCacheReuseReturnsSameHandle verifies that a cached prepare reuses
// the underlying sqlite3_stmt rather than preparing a fresh one.
func TestStmtCacheReuseReturnsSameHandle(t *testing.T) {
	d := SQLiteDriver{}
	conn, err := d.Open(":memory:?_stmt_cache_size=4")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := conn.(*SQLiteConn)
	if !c.stmtCacheEnabled {
		t.Skip("statement cache disabled on this SQLite runtime")
	}
	ctx := context.Background()

	const q = "SELECT 42"
	stmt1, err := c.prepareWithCache(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	h1 := stmt1.(*SQLiteStmt).s
	if err := stmt1.Close(); err != nil {
		t.Fatal(err)
	}

	stmt2, err := c.prepareWithCache(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	h2 := stmt2.(*SQLiteStmt).s
	if err := stmt2.Close(); err != nil {
		t.Fatal(err)
	}

	if h1 != h2 {
		t.Fatalf("expected cached prepare to reuse sqlite3_stmt handle, got %p vs %p", h1, h2)
	}
}

// TestStmtCacheSchemaChange verifies that a schema change does not let the
// cache hand back an expired statement whose captured column metadata still
// describes the old schema (issue #1447). Both same-connection DDL and DDL
// issued through a second database handle must invalidate the cache.
func TestStmtCacheSchemaChange(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "schemachange.db")
	db, err := sql.Open("sqlite3", "file:"+fn+"?_stmt_cache_size=8")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("CREATE TABLE t (a TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO t VALUES ('x')"); err != nil {
		t.Fatal(err)
	}

	queryCols := func() []string {
		rows, err := db.Query("SELECT * FROM t")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		cols, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		return cols
	}

	// Populate the cache.
	if cols := queryCols(); !reflect.DeepEqual(cols, []string{"a"}) {
		t.Fatalf("initial columns: got %v, want [a]", cols)
	}

	// Same-connection DDL.
	if _, err := db.Exec("ALTER TABLE t ADD COLUMN b TEXT"); err != nil {
		t.Fatal(err)
	}
	if cols := queryCols(); !reflect.DeepEqual(cols, []string{"a", "b"}) {
		t.Fatalf("columns after same-connection ALTER: got %v, want [a b]", cols)
	}

	// DDL through a second database handle (different connection).
	db2, err := sql.Open("sqlite3", "file:"+fn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db2.Exec("ALTER TABLE t ADD COLUMN c TEXT"); err != nil {
		db2.Close()
		t.Fatal(err)
	}
	db2.Close()
	if cols := queryCols(); !reflect.DeepEqual(cols, []string{"a", "b", "c"}) {
		t.Fatalf("columns after cross-connection ALTER: got %v, want [a b c]", cols)
	}

	// The row data must scan consistently with the new column set.
	var a string
	var b, c any
	if err := db.QueryRow("SELECT * FROM t").Scan(&a, &b, &c); err != nil {
		t.Fatal(err)
	}
	if a != "x" || b != nil || c != nil {
		t.Fatalf("row after ALTERs: got (%q, %v, %v), want (\"x\", <nil>, <nil>)", a, b, c)
	}
}

// TestStmtCacheTempSchemaChange verifies that DDL on the temp schema also
// invalidates cached statements referencing it.
func TestStmtCacheTempSchemaChange(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_stmt_cache_size=8")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("CREATE TEMP TABLE tt (a TEXT)"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM tt")
	if err != nil {
		t.Fatal(err)
	}
	rows.Close() // cached

	if _, err := db.Exec("ALTER TABLE tt ADD COLUMN b TEXT"); err != nil {
		t.Fatal(err)
	}
	rows, err = db.Query("SELECT * FROM tt")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(cols, want) {
		t.Fatalf("columns after temp ALTER: got %v, want %v", cols, want)
	}
}

// TestStmtCacheInTransaction verifies that DDL inside an explicit
// transaction is honored by queries later in the same transaction (the
// cache reports a miss there instead of probing the schema).
func TestStmtCacheInTransaction(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_stmt_cache_size=8")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("CREATE TABLE t (a TEXT)"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM t")
	if err != nil {
		t.Fatal(err)
	}
	rows.Close() // cached

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("ALTER TABLE t ADD COLUMN b TEXT"); err != nil {
		t.Fatal(err)
	}
	rows, err = tx.Query("SELECT * FROM t")
	if err != nil {
		t.Fatal(err)
	}
	cols, err := rows.Columns()
	rows.Close()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(cols, want) {
		t.Fatalf("columns inside transaction after ALTER: got %v, want %v", cols, want)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// After commit the next lookup probes again and must see the change.
	rows, err = db.Query("SELECT * FROM t")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cols, err = rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(cols, want) {
		t.Fatalf("columns after commit: got %v, want %v", cols, want)
	}
}

func cacheKeys(c *SQLiteConn) map[string]int {
	out := make(map[string]int)
	for _, s := range c.stmtCache {
		out[s.cacheKey]++
	}
	return out
}

func cacheCount(c *SQLiteConn, q string) int {
	n := 0
	for _, s := range c.stmtCache {
		if s.cacheKey == q {
			n++
		}
	}
	return n
}
