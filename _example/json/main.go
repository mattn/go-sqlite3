package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rocketlaunchr/dataframe-go"
	"github.com/rocketlaunchr/dataframe-go/imports"
)

//go:embed iris.csv
var irisData string

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	create table iris (
	    sepal_length float,
	    sepal_width  float,
	    petal_length float,
	    petal_width  float,
	    class int);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(`
	insert into iris(
	    sepal_length, sepal_width, petal_length, petal_width, class)
	    values(?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	df, err := imports.LoadFromCSV(context.Background(), strings.NewReader(irisData))
	if err != nil {
		log.Fatal(err)
	}
	it := df.ValuesIterator(dataframe.ValuesOptions{
		InitialRow:   0,
		Step:         1,
		DontReadLock: true,
	})

	df.Lock()
	for {
		row, vals, _ := it()
		if row == nil {
			break
		}
		_, err = stmt.Exec(
			vals["sepal_length"],
			vals["sepal_width"],
			vals["petal_length"],
			vals["petal_width"],
			vals["class"],
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	df.Unlock()
	tx.Commit()

	rows, err := db.Query(`select
	    json_object(
	        'sepal_length', sepal_length,
	        'sepal_width', sepal_width,
	        'petal_length', petal_length,
	        'petal_width', petal_width,
	        'class', class
	    ) from iris
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var r string
		err = rows.Scan(&r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(r)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
