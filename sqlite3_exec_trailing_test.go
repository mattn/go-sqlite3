// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestExecTrailingCommentOrSemicolonKeepsResult(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO t (id, v) VALUES (1, 'a')`); err != nil {
		t.Fatal(err)
	}

	type want struct {
		affected int64
		hasID    bool
	}
	cases := []struct {
		name  string
		query string
		args  []any
		want  want
	}{
		{name: "update", query: `UPDATE t SET v = 'b' WHERE id = 1`, want: want{affected: 1}},
		{name: "update trailing comment", query: `UPDATE t SET v = 'c' WHERE id = 1; -- mark as c`, want: want{affected: 1}},
		{name: "update args", query: `UPDATE t SET v = ? WHERE id = ?`, args: []any{"d", 1}, want: want{affected: 1}},
		{name: "update args trailing comment", query: `UPDATE t SET v = ? WHERE id = ?; -- mark as e`, args: []any{"e", 1}, want: want{affected: 1}},
		{name: "insert", query: `INSERT INTO t (v) VALUES ('f')`, want: want{affected: 1, hasID: true}},
		{name: "insert extra semicolon", query: `INSERT INTO t (v) VALUES ('g');;`, want: want{affected: 1, hasID: true}},
		{name: "insert args trailing comment", query: "INSERT INTO t (v) VALUES (?);\n-- end of script", args: []any{"h"}, want: want{affected: 1, hasID: true}},
	}

	run := func(t *testing.T, exec func(query string, args ...any) (sql.Result, error)) {
		t.Helper()
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res, err := exec(tc.query, tc.args...)
				if err != nil {
					t.Fatalf("Exec(%q) error: %v", tc.query, err)
				}
				affected, err := res.RowsAffected()
				if err != nil {
					t.Fatal(err)
				}
				if affected != tc.want.affected {
					t.Fatalf("RowsAffected=%d, want %d", affected, tc.want.affected)
				}
				id, err := res.LastInsertId()
				if err != nil {
					t.Fatal(err)
				}
				if tc.want.hasID && id == 0 {
					t.Fatalf("LastInsertId=0, want a new rowid")
				}
			})
		}
	}

	t.Run("exec", func(t *testing.T) {
		run(t, db.Exec)
	})

	// A cancellable context forces the ExecContext path (exec, not execNoArgs).
	t.Run("execContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		run(t, func(query string, args ...any) (sql.Result, error) {
			return db.ExecContext(ctx, query, args...)
		})
	})
}
