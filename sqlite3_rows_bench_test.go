// Copyright (C) 2026 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo

package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func BenchmarkPreparedRows(b *testing.B) {
	for _, count := range []int{1, 100, 1000} {
		for _, cancellable := range []bool{false, true} {
			b.Run(fmt.Sprintf("rows%d/cancellable%t", count, cancellable), func(b *testing.B) {
				db, err := sql.Open("sqlite3", ":memory:")
				if err != nil {
					b.Fatal(err)
				}
				defer db.Close()
				db.SetMaxOpenConns(1)
				_, err = db.Exec(`CREATE TABLE bench_rows (id INTEGER, label TEXT, payload BLOB, score REAL);
					WITH RECURSIVE n(x) AS (VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<1000)
					INSERT INTO bench_rows SELECT x, 'row-' || x, CAST('payload-' || x AS BLOB), x*0.5 FROM n`)
				if err != nil {
					b.Fatal(err)
				}
				stmt, err := db.Prepare("SELECT id, label, payload, score FROM bench_rows LIMIT ?")
				if err != nil {
					b.Fatal(err)
				}
				defer stmt.Close()
				ctx := context.Background()
				if cancellable {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					defer cancel()
				}
				var id int64
				var label string
				var payload []byte
				var score float64
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					rows, err := stmt.QueryContext(ctx, count)
					if err != nil {
						b.Fatal(err)
					}
					n := 0
					for rows.Next() {
						if err := rows.Scan(&id, &label, &payload, &score); err != nil {
							b.Fatal(err)
						}
						n++
					}
					if err := rows.Err(); err != nil {
						b.Fatal(err)
					}
					if err := rows.Close(); err != nil {
						b.Fatal(err)
					}
					if n != count {
						b.Fatalf("got %d rows, want %d", n, count)
					}
				}
			})
		}
	}
}
