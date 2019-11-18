// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_unlock_notify

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestUnlockNotify(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_busy_timeout=%d", tempFilename, 500)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER, status INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id, status) VALUES(1, 100)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	_, err = tx.Exec("UPDATE foo SET status = 200 WHERE id = 1")
	if err != nil {
		t.Fatal("Failed to update table:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	timer := time.NewTimer(500 * time.Millisecond)
	go func() {
		<-timer.C
		err := tx.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
		wg.Done()
	}()

	rows, err := db.Query("SELECT count(*) from foo")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			t.Fatal("Failed to Scan rows", err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal("Failed at the call to Next:", err)
	}
	wg.Wait()

}

func TestUnlockNotifyMany(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_busy_timeout=%d", tempFilename, 500)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER, status INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id, status) VALUES(1, 100)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	_, err = tx.Exec("UPDATE foo SET status = 200 WHERE id = 1")
	if err != nil {
		t.Fatal("Failed to update table:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	timer := time.NewTimer(500 * time.Millisecond)
	go func() {
		<-timer.C
		err := tx.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
		wg.Done()
	}()

	const concurrentQueries = 1000
	wg.Add(concurrentQueries)
	for i := 0; i < concurrentQueries; i++ {
		go func() {
			rows, err := db.Query("SELECT count(*) from foo")
			if err != nil {
				t.Fatal("Unable to query foo table:", err)
			}

			if rows.Next() {
				var count int
				if err := rows.Scan(&count); err != nil {
					t.Fatal("Failed to Scan rows", err)
				}
			}
			if err := rows.Err(); err != nil {
				t.Fatal("Failed at the call to Next:", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestUnlockNotifyDeadlock(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_busy_timeout=%d", tempFilename, 500)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER, status INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id, status) VALUES(1, 100)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	_, err = tx.Exec("UPDATE foo SET status = 200 WHERE id = 1")
	if err != nil {
		t.Fatal("Failed to update table:", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	timer := time.NewTimer(500 * time.Millisecond)
	go func() {
		<-timer.C
		err := tx.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		tx2, err := db.Begin()
		if err != nil {
			t.Fatal("Failed to begin transaction:", err)
		}
		defer tx2.Rollback()

		_, err = tx2.Exec("DELETE FROM foo")
		if err != nil {
			t.Fatal("Failed to delete table:", err)
		}
		err = tx2.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
		wg.Done()
	}()

	rows, err := tx.Query("SELECT count(*) from foo")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	if rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			t.Fatal("Failed to Scan rows", err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal("Failed at the call to Next:", err)
	}

	wg.Wait()
}
