// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo
// +build cgo

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNamedParams(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	create table foo (id integer, name text, extra text);
	`)
	if err != nil {
		t.Error("Failed to call db.Query:", err)
	}

	_, err = db.Exec(`insert into foo(id, name, extra) values(:id, :name, :name)`, sql.Named("name", "foo"), sql.Named("id", 1))
	if err != nil {
		t.Error("Failed to call db.Exec:", err)
	}

	row := db.QueryRow(`select id, extra from foo where id = :id and extra = :extra`, sql.Named("id", 1), sql.Named("extra", "foo"))
	if row == nil {
		t.Error("Failed to call db.QueryRow")
	}
	var id int
	var extra string
	err = row.Scan(&id, &extra)
	if err != nil {
		t.Error("Failed to db.Scan:", err)
	}
	if id != 1 || extra != "foo" {
		t.Error("Failed to db.QueryRow: not matched results")
	}
}

var (
	testTableStatements = []string{
		`DROP TABLE IF EXISTS test_table`,
		`
CREATE TABLE IF NOT EXISTS test_table (
	key1      VARCHAR(64) PRIMARY KEY,
	key_id    VARCHAR(64) NOT NULL,
	key2      VARCHAR(64) NOT NULL,
	key3      VARCHAR(64) NOT NULL,
	key4      VARCHAR(64) NOT NULL,
	key5      VARCHAR(64) NOT NULL,
	key6      VARCHAR(64) NOT NULL,
	data      BLOB        NOT NULL
);`,
	}
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func initDatabase(t *testing.T, db *sql.DB, rowCount int64) {
	for _, query := range testTableStatements {
		_, err := db.Exec(query)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := int64(0); i < rowCount; i++ {
		query := `INSERT INTO test_table
			(key1, key_id, key2, key3, key4, key5, key6, data)
			VALUES
			(?, ?, ?, ?, ?, ?, ?, ?);`
		args := []interface{}{
			randStringBytes(50),
			fmt.Sprint(i),
			randStringBytes(50),
			randStringBytes(50),
			randStringBytes(50),
			randStringBytes(50),
			randStringBytes(50),
			randStringBytes(50),
			randStringBytes(2048),
		}
		_, err := db.Exec(query, args...)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestShortTimeout(t *testing.T) {
	srcTempFilename := TempFilename(t)
	defer os.Remove(srcTempFilename)

	db, err := sql.Open("sqlite3", srcTempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	initDatabase(t, db, 100)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()
	query := `SELECT key1, key_id, key2, key3, key4, key5, key6, data
		FROM test_table
		ORDER BY key2 ASC`
	_, err = db.QueryContext(ctx, query)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatal(err)
	}
	if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
		t.Fatal(ctx.Err())
	}
}

func TestExecContextCancel(t *testing.T) {
	srcTempFilename := TempFilename(t)
	defer os.Remove(srcTempFilename)

	db, err := sql.Open("sqlite3", srcTempFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	ts := time.Now()
	initDatabase(t, db, 1000)
	spent := time.Since(ts)
	const minTestTime = 100 * time.Millisecond
	if spent < minTestTime {
		t.Skipf("test will be too racy (spent=%s < min=%s) as ExecContext below will be too fast.",
			spent.String(), minTestTime.String(),
		)
	}

	// expected to be extremely slow query
	q := `
INSERT INTO test_table (key1, key_id, key2, key3, key4, key5, key6, data)
SELECT t1.key1 || t2.key1, t1.key_id || t2.key_id, t1.key2 || t2.key2, t1.key3 || t2.key3, t1.key4 || t2.key4, t1.key5 || t2.key5, t1.key6 || t2.key6, t1.data || t2.data
FROM test_table t1 LEFT OUTER JOIN test_table t2`
	// expect query above take ~ same time as setup above
	// This is racy: the context must be valid so sql/db.ExecContext calls the sqlite3 driver.
	// It starts the query, the context expires, then calls sqlite3_interrupt
	ctx, cancel := context.WithTimeout(context.Background(), minTestTime/2)
	defer cancel()
	ts = time.Now()
	r, err := db.ExecContext(ctx, q)
	// racy check
	if r != nil {
		n, err := r.RowsAffected()
		t.Logf("query should not have succeeded: rows=%d; err=%v; duration=%s",
			n, err, time.Since(ts).String())
	}
	if err != context.DeadlineExceeded {
		t.Fatal(err, ctx.Err())
	}
}

func TestQueryRowContextCancel(t *testing.T) {
	srcTempFilename := TempFilename(t)
	defer os.Remove(srcTempFilename)

	db, err := sql.Open("sqlite3", srcTempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	initDatabase(t, db, 100)

	const query = `SELECT key_id FROM test_table ORDER BY key2 ASC`
	var keyID string
	unexpectedErrors := make(map[string]int)
	for i := 0; i < 10000; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		row := db.QueryRowContext(ctx, query)

		cancel()
		// it is fine to get "nil" as context cancellation can be handled with delay
		if err := row.Scan(&keyID); err != nil && err != context.Canceled {
			if err.Error() == "sql: Rows are closed" {
				// see https://github.com/golang/go/issues/24431
				// fixed in 1.11.1 to properly return context error
				continue
			}
			unexpectedErrors[err.Error()]++
		}
	}
	for errText, count := range unexpectedErrors {
		t.Error(errText, count)
	}
}

func TestQueryRowContextCancelParallel(t *testing.T) {
	srcTempFilename := TempFilename(t)
	defer os.Remove(srcTempFilename)

	db, err := sql.Open("sqlite3", srcTempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	initDatabase(t, db, 100)

	const query = `SELECT key_id FROM test_table ORDER BY key2 ASC`
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var keyID string
			for j := 0; j < 100; j++ {
				ctx, cancel := context.WithCancel(context.Background())
				row := db.QueryRowContext(ctx, query)
				cancel()
				_ = row.Scan(&keyID)
			}
		}()
	}
	wg.Wait()
}

func TestBeginTxCancel(t *testing.T) {
	srcTempFilename := TempFilename(t)
	defer os.Remove(srcTempFilename)

	db, err := sql.Open("sqlite3", srcTempFilename)
	if err != nil {
		t.Fatal(err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	defer db.Close()
	initDatabase(t, db, 100)

	// create several go-routines to expose racy issue
	for i := 0; i < 1000; i++ {
		func() {
			ctx, cancel := context.WithCancel(context.Background())
			conn, err := db.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := conn.Close(); err != nil {
					t.Error(err)
				}
			}()

			err = conn.Raw(func(driverConn any) error {
				d, ok := driverConn.(driver.ConnBeginTx)
				if !ok {
					t.Fatal("unexpected: wrong type")
				}
				// checks that conn.Raw can be used to get *SQLiteConn
				if _, ok = driverConn.(*SQLiteConn); !ok {
					t.Fatalf("conn.Raw() driverConn type=%T, expected *SQLiteConn", driverConn)
				}

				go cancel() // make it cancel concurrently with exec("BEGIN");
				tx, err := d.BeginTx(ctx, driver.TxOptions{})
				switch err {
				case nil:
					switch err := tx.Rollback(); err {
					case nil, sql.ErrTxDone:
					default:
						return err
					}
				case context.Canceled:
				default:
					// must not fail with "cannot start a transaction within a transaction"
					return err
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
}

func TestStmtReadonly(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE t (count INT)")
	if err != nil {
		t.Fatal(err)
	}

	isRO := func(query string) bool {
		c, err := db.Conn(context.Background())
		if err != nil {
			return false
		}

		var ro bool
		c.Raw(func(dc any) error {
			stmt, err := dc.(*SQLiteConn).Prepare(query)
			if err != nil {
				return err
			}
			if stmt == nil {
				return errors.New("stmt is nil")
			}
			ro = stmt.(*SQLiteStmt).Readonly()
			return nil
		})
		return ro // On errors ro will remain false.
	}

	if !isRO(`select * from t`) {
		t.Error("select not seen as read-only")
	}
	if isRO(`insert into t values (1), (2)`) {
		t.Error("insert seen as read-only")
	}
}
