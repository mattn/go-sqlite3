package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/mattn/go-sqlite3"
)

func main() {
	sql.Register("sqlite3_with_wal_hook_example",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conn.RegisterWalHook(func(dbName string, nPages int) int {
					if nPages >= 1 {
						if _, err := conn.Exec("PRAGMA wal_checkpoint(TRUNCATE);", nil); err != nil {
							log.Fatal(err)
						}
					}
					return sqlite3.SQLITE_OK
				})
				return nil
			},
		})
	defer os.Remove("./foo.db")

	db, err := sql.Open("sqlite3_with_wal_hook_example", "./foo.db?_journal=WAL")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table foo(id int, value text)")
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into foo(id, value) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 100; i++ {
		if _, err := stmt.Exec(i, "value"); err != nil {
			log.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	var busy, log_, checkpointed int
	err = db.QueryRow("PRAGMA wal_checkpoint(PASSIVE);").Scan(&busy, &log_, &checkpointed)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("busy=%d log=%d checkpointed=%d\n", busy, log_, checkpointed) // busy=0 log=0 checkpointed=0
}
