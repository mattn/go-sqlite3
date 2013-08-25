package main

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	const (
		use_hook = true
		load_query = "SELECT load_extension('sqlite3_mod_regexp.dll')"
	)

	sql.Register("sqlite3_with_extensions",
		&sqlite3.SQLiteDriver{
			EnableLoadExtension: true,
			ConnectHook: func(c *sqlite3.SQLiteConn) error {
				if use_hook {
					stmt, err := c.Prepare(load_query)
					if err != nil {
						return err
					}

					_, err = stmt.Exec(nil)
					if err != nil {
						return err
					}

					return stmt.Close()
				}
				return nil
			},
		})

	db, err := sql.Open("sqlite3_with_extensions", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if !use_hook {
		if _, err = db.Exec(load_query); err != nil {
			log.Fatal(err)
		}
	}

	// Force db to make a new connection in pool
	// by putting the original in a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	// New connection works (hopefully!)
	rows, err := db.Query("select 'hello world' where 'hello world' regexp '^hello.*d$'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var helloworld string
		rows.Scan(&helloworld)
		fmt.Println(helloworld)
	}
}
