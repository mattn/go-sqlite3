// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_unlock_notify
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
		defer wg.Done()
		<-timer.C
		err := tx.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
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
		defer wg.Done()
		<-timer.C
		err := tx.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
	}()

	const concurrentQueries = 1000
	wg.Add(concurrentQueries)
	for i := 0; i < concurrentQueries; i++ {
		go func() {
			defer wg.Done()
			rows, err := db.Query("SELECT count(*) from foo")
			if err != nil {
				t.Error("Unable to query foo table:", err)
				return
			}

			if rows.Next() {
				var count int
				if err := rows.Scan(&count); err != nil {
					t.Error("Failed to Scan rows", err)
					return
				}
				if count != 1 {
					t.Errorf("count=%d want=%d", count, 1)
				}
			}
			if err := rows.Err(); err != nil {
				t.Error("Failed at the call to Next:", err)
				return
			}
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
		defer wg.Done()
		<-timer.C
		err := tx.Commit()
		if err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
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
