// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build go1.8 && cgo
// +build go1.8,cgo

package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
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
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	defer db.Close()
	initDatabase(t, db, 100)

	const query = `SELECT key_id FROM test_table ORDER BY key2 ASC`
	wg := sync.WaitGroup{}
	defer wg.Wait()

	testCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var keyID string
			for {
				select {
				case <-testCtx.Done():
					return
				default:
				}
				ctx, cancel := context.WithCancel(context.Background())
				row := db.QueryRowContext(ctx, query)

				cancel()
				_ = row.Scan(&keyID) // see TestQueryRowContextCancel
			}
		}()
	}

	var keyID string
	for i := 0; i < 10000; i++ {
		// note that testCtx is not cancelled during query execution
		row := db.QueryRowContext(testCtx, query)

		if err := row.Scan(&keyID); err != nil {
			t.Fatal(i, err)
		}
	}
}

// Test that we can successfully interrupt a long running query when
// the context is canceled. The previous two QueryRowContext tests
// only test that we handle a previously cancelled context and thus
// do not call sqlite3_interrupt.
func TestQueryRowContextCancelInterrupt(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test that we have the unixepoch function and if not skip the test.
	if _, err := db.Exec(`SELECT unixepoch(datetime(100000, 'unixepoch', 'localtime'))`); err != nil {
		libVersion, libVersionNumber, sourceID := Version()
		if strings.Contains(err.Error(), "no such function: unixepoch") {
			t.Skip("Skipping the 'unixepoch' function is not implemented in "+
				"this version of sqlite3:", libVersion, libVersionNumber, sourceID)
		}
		t.Fatal(err)
	}

	const createTableStmt = `
	CREATE TABLE timestamps (
		ts TIMESTAMP NOT NULL
	);`
	if _, err := db.Exec(createTableStmt); err != nil {
		t.Fatal(err)
	}

	stmt, err := db.Prepare(`INSERT INTO timestamps VALUES (?);`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	// Computationally expensive query that consumes many rows. This is needed
	// to test cancellation because queries are not interrupted immediately.
	// Instead, queries are only halted at certain checkpoints where the
	// sqlite3.isInterrupted is checked and true.
	queryStmt := `
	SELECT
		SUM(unixepoch(datetime(ts + 10, 'unixepoch', 'localtime'))) AS c1,
		SUM(unixepoch(datetime(ts + 20, 'unixepoch', 'localtime'))) AS c2,
		SUM(unixepoch(datetime(ts + 30, 'unixepoch', 'localtime'))) AS c3,
		SUM(unixepoch(datetime(ts + 40, 'unixepoch', 'localtime'))) AS c4
	FROM
		timestamps
	WHERE datetime(ts, 'unixepoch', 'localtime')
	LIKE
		?;`

	query := func(t *testing.T, timeout time.Duration) (int, error) {
		// Create a complicated pattern to match timestamps
		const pattern = "%2%0%2%4%-%-%:%:%"
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		rows, err := db.QueryContext(ctx, queryStmt, pattern)
		if err != nil {
			return 0, err
		}
		var count int
		for rows.Next() {
			var n int64
			if err := rows.Scan(&n, &n, &n, &n); err != nil {
				return count, err
			}
			count++
		}
		return count, rows.Err()
	}

	average := func(n int, fn func()) time.Duration {
		start := time.Now()
		for i := 0; i < n; i++ {
			fn()
		}
		return time.Since(start) / time.Duration(n)
	}

	createRows := func(n int) {
		t.Logf("Creating %d rows", n)
		if _, err := db.Exec(`DELETE FROM timestamps; VACUUM;`); err != nil {
			t.Fatal(err)
		}
		ts := time.Date(2024, 6, 6, 8, 9, 10, 12345, time.UTC).Unix()
		rr := rand.New(rand.NewSource(1234))
		for i := 0; i < n; i++ {
			if _, err := stmt.Exec(ts + rr.Int63n(10_000) - 5_000); err != nil {
				t.Fatal(err)
			}
		}
	}

	const TargetRuntime = 200 * time.Millisecond
	const N = 5_000 // Number of rows to insert at a time

	// Create enough rows that the query takes ~200ms to run.
	start := time.Now()
	createRows(N)
	baseAvg := average(4, func() {
		if _, err := query(t, time.Hour); err != nil {
			t.Fatal(err)
		}
	})
	t.Log("Base average:", baseAvg)
	rowCount := N * (int(TargetRuntime/baseAvg) + 1)
	createRows(rowCount)
	t.Log("Table setup time:", time.Since(start))

	// Set the timeout to 1/10 of the average query time.
	avg := average(2, func() {
		n, err := query(t, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			t.Fatal("scanned zero rows")
		}
	})
	// Guard against the timeout being too short to reliably test.
	if avg < TargetRuntime/2 {
		t.Fatalf("Average query runtime should be around %s got: %s ",
			TargetRuntime, avg)
	}
	timeout := (avg / 10).Round(100 * time.Microsecond)
	t.Logf("Average: %s Timeout: %s", avg, timeout)

	for i := 0; i < 10; i++ {
		tt := time.Now()
		n, err := query(t, timeout)
		if !errors.Is(err, context.DeadlineExceeded) {
			fn := t.Errorf
			if err != nil {
				fn = t.Fatalf
			}
			fn("expected error %v got %v", context.DeadlineExceeded, err)
		}
		d := time.Since(tt)
		t.Logf("%d: rows: %d duration: %s", i, n, d)
		if d > timeout*4 {
			t.Errorf("query was cancelled after %s but did not abort until: %s", timeout, d)
		}
	}
}

func TestExecCancel(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("create table foo (id integer primary key)"); err != nil {
		t.Fatal(err)
	}

	for n := 0; n < 100; n++ {
		ctx, cancel := context.WithCancel(context.Background())
		_, err = db.ExecContext(ctx, "insert into foo (id) values (?)", n)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func doTestOpenContext(t *testing.T, option string) (string, error) {
	tempFilename := TempFilename(t)
	url := tempFilename + option

	defer func() {
		err := os.Remove(tempFilename)
		if err != nil {
			t.Error("temp file remove error:", err)
		}
	}()

	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return "Failed to open database:", err
	}

	defer func() {
		err = db.Close()
		if err != nil {
			t.Error("db close error:", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		return "ping error:", err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "drop table foo")
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "create table foo (id integer)")
	cancel()
	if err != nil {
		return "Failed to create table:", err
	}

	if stat, err := os.Stat(tempFilename); err != nil || stat.IsDir() {
		return "Failed to create ./foo.db", nil
	}

	return "", nil
}

func TestOpenContext(t *testing.T) {
	cases := map[string]bool{
		"":                   true,
		"?_txlock=immediate": true,
		"?_txlock=deferred":  true,
		"?_txlock=exclusive": true,
		"?_txlock=bogus":     false,
	}
	for option, expectedPass := range cases {
		result, err := doTestOpenContext(t, option)
		if result == "" {
			if !expectedPass {
				errmsg := fmt.Sprintf("_txlock error not caught at dbOpen with option: %s", option)
				t.Fatal(errmsg)
			}
		} else if expectedPass {
			if err == nil {
				t.Fatal(result)
			} else {
				t.Fatal(result, err)
			}
		}
	}
}

func TestFileCopyTruncate(t *testing.T) {
	var err error
	tempFilename := TempFilename(t)

	defer func() {
		err = os.Remove(tempFilename)
		if err != nil {
			t.Error("temp file remove error:", err)
		}
	}()

	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("open error:", err)
	}

	defer func() {
		err = db.Close()
		if err != nil {
			t.Error("db close error:", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		t.Fatal("ping error:", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "drop table foo")
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "create table foo (id integer)")
	cancel()
	if err != nil {
		t.Fatal("create table error:", err)
	}

	// copy db to new file
	var data []byte
	data, err = ioutil.ReadFile(tempFilename)
	if err != nil {
		t.Fatal("read file error:", err)
	}

	var f *os.File
	f, err = os.Create(tempFilename + "-db-copy")
	if err != nil {
		t.Fatal("create file error:", err)
	}

	defer func() {
		err = os.Remove(tempFilename + "-db-copy")
		if err != nil {
			t.Error("temp file moved remove error:", err)
		}
	}()

	_, err = f.Write(data)
	if err != nil {
		f.Close()
		t.Fatal("write file error:", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal("close file error:", err)
	}

	// truncate current db file
	f, err = os.OpenFile(tempFilename, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatal("open file error:", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal("close file error:", err)
	}

	// test db after file truncate
	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		t.Fatal("ping error:", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "drop table foo")
	cancel()
	if err == nil {
		t.Fatal("drop table no error")
	}

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "create table foo (id integer)")
	cancel()
	if err != nil {
		t.Fatal("create table error:", err)
	}

	err = db.Close()
	if err != nil {
		t.Error("db close error:", err)
	}

	// test copied file
	db, err = sql.Open("sqlite3", tempFilename+"-db-copy")
	if err != nil {
		t.Fatal("open error:", err)
	}

	defer func() {
		err = db.Close()
		if err != nil {
			t.Error("db close error:", err)
		}
	}()

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		t.Fatal("ping error:", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "drop table foo")
	cancel()
	if err != nil {
		t.Fatal("drop table error:", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 55*time.Second)
	_, err = db.ExecContext(ctx, "create table foo (id integer)")
	cancel()
	if err != nil {
		t.Fatal("create table error:", err)
	}
}
