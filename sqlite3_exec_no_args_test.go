//go:build cgo

package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func TestExecNoArgs(t *testing.T) {
	tests := []struct {
		name  string
		query string
		args  []any
		want  int
		rows  int
		nulls int
	}{
		{name: "positional", query: "INSERT INTO t(v) VALUES(?)", want: 1},
		{name: "numbered", query: "INSERT INTO t(v) VALUES(?3)", want: 3},
		{name: "named", query: "INSERT INTO t(v) VALUES(:v)", want: 1},
		{name: "repeated_named", query: "INSERT INTO t(v) VALUES(:v || :v)", want: 1},
		{name: "multiple", query: "INSERT INTO t(v) VALUES(? || ?)", want: 2},
		{name: "later_statement", query: "INSERT INTO t(v) VALUES('first'); INSERT INTO t(v) VALUES(?); INSERT INTO t(v) VALUES('last')", want: 1, rows: 1},
		{name: "first_statement", query: "INSERT INTO t(v) VALUES(?); INSERT INTO t(v) VALUES('last')", want: 1},
		{name: "explicit_nil", query: "INSERT INTO t(v) VALUES(?)", args: []any{nil}, rows: 1, nulls: 1},
		{name: "literal", query: "INSERT INTO t(v) VALUES('? :v ?3')", rows: 1},
		{name: "comment", query: "/* ? :v */ INSERT INTO t(v) VALUES('ok'); -- ?3\n", rows: 1},
		{name: "batch", query: "INSERT INTO t(v) VALUES('first'); INSERT INTO t(v) VALUES('last')", rows: 2},
		{name: "empty", query: ""},
		{name: "comments_only", query: " ; /* ? */ -- :v\n"},
	}
	for _, tc := range tests {
		for _, cancellable := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cancellable=%t", tc.name, cancellable), func(t *testing.T) {
				db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				if _, err := db.Exec("CREATE TABLE t(v)"); err != nil {
					t.Fatal(err)
				}
				ctx := context.Background()
				if cancellable {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					defer cancel()
				}
				_, err = db.ExecContext(ctx, tc.query, tc.args...)
				if tc.want > 0 {
					want := fmt.Sprintf("not enough args to execute query: want %d got 0", tc.want)
					if err == nil || err.Error() != want {
						t.Errorf("ExecContext error = %v, want %q", err, want)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				var rows, nulls int
				if err := db.QueryRow("SELECT count(*), count(*) - count(v) FROM t").Scan(&rows, &nulls); err != nil {
					t.Fatal(err)
				}
				if rows != tc.rows || nulls != tc.nulls {
					t.Errorf("stored rows/nulls = %d/%d, want %d/%d", rows, nulls, tc.rows, tc.nulls)
				}
			})
		}
	}
}

func TestExecNoArgsResult(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	res, err := db.Exec("CREATE TABLE t(v); INSERT INTO t VALUES('first'); INSERT INTO t VALUES('last')")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := res.LastInsertId(); err != nil || id != 2 {
		t.Errorf("LastInsertId = %d, %v; want 2, nil", id, err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows != 1 {
		t.Errorf("RowsAffected = %d, %v; want 1, nil", rows, err)
	}
}

func TestExecNoArgsSQLiteError(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("INSERT INTO missing_table VALUES(1)"); err == nil || err.Error() != "no such table: missing_table" {
		t.Fatalf("Exec prepare error = %v, want missing table error", err)
	}
	if _, err := db.Exec("CREATE TABLE t(v UNIQUE); INSERT INTO t VALUES(1)"); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO t VALUES(1)")
	var sqliteErr Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code != ErrConstraint || sqliteErr.ExtendedCode != ErrConstraintUnique {
		t.Fatalf("Exec error = %#v, want unique constraint error", err)
	}
	// The connection must remain usable after both argument and SQLite errors.
	if _, err := db.Exec("INSERT INTO t VALUES(?)"); err == nil {
		t.Fatal("missing argument was accepted")
	}
	if _, err := db.Exec("INSERT INTO t VALUES(2)"); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkExecNoArgs(b *testing.B) {
	for _, query := range []string{"SELECT 1", "UPDATE t SET v = v + 1", "SELECT 1; SELECT 2"} {
		b.Run(query, func(b *testing.B) {
			db, err := sql.Open("sqlite3", ":memory:")
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()
			if _, err := db.Exec("CREATE TABLE t(v); INSERT INTO t VALUES(0)"); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := db.Exec(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
