// +build sqlite_math_functions

package sqlite3

import (
	"database/sql"
	"testing"
)

func TestMathFunctions(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	queries := []string{
		`SELECT acos(1)`,
		`SELECT log(10, 100)`,
		`SELECT power(2, 2)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatal("Failed to call math function:", err)
		}
	}
}
