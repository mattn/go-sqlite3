package sqlite

import (
	"fmt"
	"testing"
	"exp/sql"
	"os"
)

func TestOpen(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	if stat, err := os.Stat("./foo.db"); err != nil || stat.IsDirectory() {
		t.Errorf("Failed to create ./foo.db")
	}
}

func TestInsert(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	_, err = db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	rows.Next()

	var result int
	rows.Scan(&result)
	if result != 123 {
		t.Errorf("Fetched %q; expected %q", 123, result)
	}
}

func TestUpdate(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	_, err = db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	_, err = db.Exec("update foo set id = 234")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	rows.Next()

	var result int
	rows.Scan(&result)
	if result != 234 {
		t.Errorf("Fetched %q; expected %q", 234, result)
	}
}

func TestDelete(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	_, err = db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	_, err = db.Exec("delete from foo where id = 123")
	if err != nil {
		t.Errorf("Failed to create table")
		return
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		t.Errorf("Fetched row but expected not rows")
	}
}
