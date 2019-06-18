// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build go1.8

package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
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
