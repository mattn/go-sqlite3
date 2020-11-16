// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build go1.13,cgo

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"testing"
)

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

			err = conn.Raw(func(driverConn interface{}) error {
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
