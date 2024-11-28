package main

import (
	"database/sql"

	_ "github.com/charlievieth/go-sqlite3"
)

func main() {
	for _, driver := range sql.Drivers() {
		println(driver)
	}
}
