package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

func main() {
	log.Printf("args: %d\n", len(os.Args))
	if len(os.Args) < 2 {
		log.Fatal("not enough arguments")
	}
	path := os.Args[1]
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if len(os.Args) < 3 {
		dumbShell(db)
	} else {
		query := os.Args[2]
		runQuery(db, query)
	}
}

func dumbShell(db *sql.DB) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("sql> ")
		var query string
		query, _ = reader.ReadString('\n')
		//fmt.Println(query)
		query = strings.TrimSpace(query)
		if len(query) > 0 {
			runQuery(db, query)
		}
	}
}

func runQuery(db *sql.DB, query string) (err error) {
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		return err
	}
	cols, err := rows.ColumnTypes()
	if err != nil {
		log.Println(err)
		return err
	}
	// http://go-database-sql.org/varcols.html
	vals := make([]interface{}, len(cols))
	for i, types := range cols {
		vals[i] = new(sql.RawBytes)
		fmt.Printf("%-50s\t", types.Name())
		/* debugging...
		        length, ok := types.Length()
				if !ok {
					length = 0
				}
				fmt.Printf("%s[%s]\t", types.DatabaseTypeName(), length)
		*/
	}
	fmt.Printf("\n")
	for rows.Next() {
		if err := rows.Scan(vals...); err != nil {
			log.Println(err)
			return err
		}
		// Now you can check each element of vals for nil-ness,
		// and you can use type introspection and type assertions
		// to fetch the column into a typed variable.
		for i, _ := range vals {
			if vals[i] == nil {
				fmt.Printf("%-50s\t", "NULL")
			} else {
				fmt.Printf("%-50s\t", vals[i])
			}
		}
		fmt.Printf("\n")
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}
	return err
}
