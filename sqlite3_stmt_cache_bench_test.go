// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo
// +build cgo

package sqlite3

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"
)

// BenchmarkStmtCache measures the stmt cache hit / miss / eviction paths by
// cycling through a fixed set of queries under various cache sizes. It is
// intended for comparing cache behavior changes, not for absolute numbers.
func BenchmarkStmtCache(b *testing.B) {
	cases := []struct {
		name      string
		cacheSize int
		keyCount  int
	}{
		{"off", 0, 1},                 // baseline: no cache
		{"size4_keys1_hit", 4, 1},     // trivial hit path
		{"size4_keys4_hit", 4, 4},     // all queries fit, always hit
		{"size4_keys8_evict", 4, 8},   // working set > cache: miss + eviction
		{"size16_keys8_hit", 16, 8},   // all queries fit in larger cache
		{"size16_keys32_evict", 16, 32}, // working set >> cache
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			dsn := ":memory:"
			if tc.cacheSize > 0 {
				dsn = fmt.Sprintf(":memory:?_stmt_cache_size=%d", tc.cacheSize)
			}
			d := SQLiteDriver{}
			conn, err := d.Open(dsn)
			if err != nil {
				b.Fatal(err)
			}
			defer conn.Close()
			c := conn.(*SQLiteConn)

			queries := make([]string, tc.keyCount)
			for i := range queries {
				// Distinct literal forces a distinct prepared statement.
				queries[i] = fmt.Sprintf("SELECT %d", i+1)
			}

			ctx := context.Background()
			// Warm up: exercise each query at least once so the cache (if any)
			// reaches steady state before timing begins.
			for _, q := range queries {
				rows, err := c.query(ctx, q, nil)
				if err != nil {
					b.Fatal(err)
				}
				drainRows(b, rows)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q := queries[i%len(queries)]
				rows, err := c.query(ctx, q, nil)
				if err != nil {
					b.Fatal(err)
				}
				drainRows(b, rows)
			}
		})
	}
}

func drainRows(b *testing.B, rows driver.Rows) {
	b.Helper()
	dest := make([]driver.Value, len(rows.Columns()))
	for {
		if err := rows.Next(dest); err != nil {
			break
		}
	}
	rows.Close()
}
