package sqlite3_test

import (
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/CanonicalLtd/go-sqlite3"
)

// Return a SQLiteConn opened against the given DSN.
func newSQLiteConn(name string) *sqlite3.SQLiteConn {
	driver := &sqlite3.SQLiteDriver{}
	conn, err := driver.Open(name)
	if err != nil {
		panic(fmt.Sprintf("can't open connection to %s: %v", name, err))
	}
	return conn.(*sqlite3.SQLiteConn)
}

// Return a SQLiteConn opened against a temporary database filename,
// along with a cleanup function that can be used to purge all
// temporary files.
func newFileSQLiteConn() (*sqlite3.SQLiteConn, func()) {
	dir, err := ioutil.TempDir("", "sqlite3-database-")
	if err != nil {
		panic(fmt.Sprintf("can't create database temp file: %v", err))
	}

	conn := newSQLiteConn(filepath.Join(dir, "test.db"))

	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(fmt.Sprintf("can't remove dir with temp database: %v", err))
		}
	}

	return conn, cleanup
}

// Return a SQLiteConn opened against the :memory" backend.
func newMemorySQLiteConn() *sqlite3.SQLiteConn {
	return newSQLiteConn(":memory:")
}

// Wrapper around SQLiteConn.Exec which panics in case of error.
func mustExec(conn *sqlite3.SQLiteConn, query string, args []driver.Value) driver.Result {
	result, err := conn.Exec(query, args)
	if err != nil {
		panic(fmt.Sprintf("exec %s [%v] failed: %v", query, args, err))
	}
	return result
}

// Wrapper around SQLiteConn.Query which panics in case of error.
func mustQuery(conn *sqlite3.SQLiteConn, query string, args []driver.Value) driver.Rows {
	rows, err := conn.Query(query, args)
	if err != nil {
		panic(fmt.Sprintf("query %s [%v] failed: %v", query, args, err))
	}
	return rows
}
