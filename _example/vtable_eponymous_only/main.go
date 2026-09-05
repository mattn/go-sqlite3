package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/mattn/go-sqlite3"
)

func main() {
	sql.Register("sqlite3_with_extensions", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.CreateModule("series", &seriesModule{})
		},
	})
	db, err := sql.Open("sqlite3_with_extensions", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, query := range []string{
		"select value from series where value between 3 and 5",
		"select value from series where start = 10 and stop = 20 and step = 5",
		"select value from series where start = 3 and stop = -3 and step = -2",
	} {
		fmt.Println(query)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var value int
			if err := rows.Scan(&value); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("value: %d\n", value)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
		rows.Close()
	}
}
