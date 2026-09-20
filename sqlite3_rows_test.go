//go:build cgo

package sqlite3

import (
	"context"
	"database/sql/driver"
	"io"
	"reflect"
	"testing"
)

func TestRowsStepValues(t *testing.T) {
	for _, cancellable := range []bool{false, true} {
		name := "background"
		if cancellable {
			name = "cancellable"
		}
		t.Run(name, func(t *testing.T) {
			conn := openContextTestConn(t)
			ctx := context.Background()
			if cancellable {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				defer cancel()
			}
			stmt, err := conn.PrepareContext(ctx, `SELECT 'first', x'0102', 1, 1.5 WHERE :include
				UNION ALL SELECT x'0304', 'second', NULL, 2
				UNION ALL SELECT '', x'', 3.5, NULL`)
			if err != nil {
				t.Fatal(err)
			}
			defer stmt.Close()
			want := [][]driver.Value{
				{"first", []byte{1, 2}, int64(1), float64(1.5)},
				{[]byte{3, 4}, "second", nil, int64(2)},
				{"", []byte{}, float64(3.5), nil},
			}
			// Alternate fused and generic binding with the same buffers.
			for run := 0; run < 3; run++ {
				args := []driver.NamedValue{{Ordinal: 1, Value: int64(1)}}
				if run == 1 {
					args[0].Name = "include"
				}
				rows, err := stmt.(driver.StmtQueryContext).QueryContext(ctx, args)
				if err != nil {
					t.Fatal(err)
				}
				var got [][]driver.Value
				for range want {
					values := make([]driver.Value, 4)
					if err := rows.Next(values); err != nil {
						t.Fatal(err)
					}
					got = append(got, values)
				}
				if err := rows.Next(make([]driver.Value, 4)); err != io.EOF {
					t.Fatalf("Next = %v, want EOF", err)
				}
				if err := rows.Close(); err != nil {
					t.Fatal(err)
				}
				// Retained strings and blobs must survive subsequent steps and reset.
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			}
		})
	}
}
