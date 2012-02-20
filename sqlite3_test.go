package sqlite

import (
	"database/sql"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Errorf("Failed to open database:", err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}

	if stat, err := os.Stat("./foo.db"); err != nil || stat.IsDir() {
		t.Errorf("Failed to create ./foo.db")
	}
}

func TestInsert(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Errorf("Failed to open database:", err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}

	res, err := db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Errorf("Failed to insert record:", err)
		return
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		t.Errorf("Expected %d for affected rows, but %d:", 1, affected)
		return
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Errorf("Failed to select records:", err)
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
		t.Errorf("Failed to open database:", err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}

	res, err := db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Errorf("Failed to insert record:", err)
		return
	}
	expected, err := res.LastInsertId()
	if err != nil {
		t.Errorf("Failed to get LastInsertId:", err)
		return
	}
	affected, _ := res.RowsAffected()
	if err != nil {
		t.Errorf("Failed to get RowsAffected:", err)
		return
	}
	if affected != 1 {
		t.Errorf("Expected %d for affected rows, but %d:", 1, affected)
		return
	}

	res, err = db.Exec("update foo set id = 234")
	if err != nil {
		t.Errorf("Failed to update record:", err)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		t.Errorf("Failed to get LastInsertId:", err)
		return
	}
	if expected != lastId {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastId)
	}
	affected, _ = res.RowsAffected()
	if err != nil {
		t.Errorf("Failed to get RowsAffected:", err)
		return
	}
	if affected != 1 {
		t.Errorf("Expected %d for affected rows, but %d:", 1, affected)
		return
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Errorf("Failed to select records:", err)
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
		t.Errorf("Failed to select records:", err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}

	res, err := db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Errorf("Failed to insert record:", err)
		return
	}
	expected, err := res.LastInsertId()
	if err != nil {
		t.Errorf("Failed to get LastInsertId:", err)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		t.Errorf("Failed to get RowsAffected:", err)
		return
	}
	if affected != 1 {
		t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
	}

	res, err = db.Exec("delete from foo where id = 123")
	if err != nil {
		t.Errorf("Failed to delete record:", err)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		t.Errorf("Failed to get LastInsertId:", err)
		return
	}
	if expected != lastId {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastId)
	}
	affected, err = res.RowsAffected()
	if err != nil {
		t.Errorf("Failed to get RowsAffected:", err)
		return
	}
	if affected != 1 {
		t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Errorf("Failed to select records:", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		t.Errorf("Fetched row but expected not rows")
	}
}

func TestBooleanRoundtrip(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Errorf("Tailed to open database:", err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, value BOOL)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}

	_, err = db.Exec("INSERT INTO foo(id, value) VALUES(1, ?)", true)
	if err != nil {
		t.Errorf("Failed to insert true value:", err)
		return
	}

	_, err = db.Exec("INSERT INTO foo(id, value) VALUES(2, ?)", false)
	if err != nil {
		t.Errorf("Failed to insert false value:", err)
		return
	}

	rows, err := db.Query("SELECT id, value FROM foo")
	if err != nil {
		t.Errorf("Unable to query foo table:", err)
		return
	}

	for rows.Next() {
		var id int
		var value bool

		if err := rows.Scan(&id, &value); err != nil {
			t.Errorf("Unable to scan results:", err)
			continue
		}

		if id == 1 && !value {
			t.Errorf("Value for id 1 should be true, not false")

		} else if id == 2 && value {
			t.Errorf("Value for id 2 should be false, not true")
		}
	}
}
