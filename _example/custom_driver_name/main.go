package main

import (
	"database/sql"

	_ "github.com/fhir-fli/go-sqlite3-sqlcipher"
)

func main() {
	for _, driver := range sql.Drivers() {
		println(driver)
	}
}
