package main

import (
	"database/sql"
	"log"

	_ "github.com/fhir-fli/go-sqlite3-sqlcipher"
)

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "file:test_encrypted.db?mode=rwc&_pragma_key=test_secret")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a table to test
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, value TEXT);")
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Database created and encrypted successfully!")
}
