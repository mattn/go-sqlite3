package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	db, err := sql.Open("sqlite3_with_extensions", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("select load_extension('sqlite3_mod_regexp.dll')")
	if err != nil {
		log.Fatal(err)
	}

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
