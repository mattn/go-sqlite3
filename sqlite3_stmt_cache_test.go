// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo
// +build cgo

package sqlite3

import (
	"context"
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
