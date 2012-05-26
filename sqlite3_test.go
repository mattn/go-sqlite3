package sqlite

import (
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"
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
		t.Errorf("Failed to open database:", err)
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

func TestTimestamp(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Errorf("Failed to open database:", err)
		return
	}
	defer os.Remove("./foo.db")

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, ts timeSTAMP)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}

	timestamp1 := time.Date(2012, time.April, 6, 22, 50, 0, 0, time.UTC)
	_, err = db.Exec("INSERT INTO foo(id, ts) VALUES(1, ?)", timestamp1)
	if err != nil {
		t.Errorf("Failed to insert timestamp:", err)
		return
	}

	timestamp2 := time.Date(2012, time.April, 6, 23, 22, 0, 0, time.UTC)
	_, err = db.Exec("INSERT INTO foo(id, ts) VALUES(2, ?)", timestamp2.Unix())
	if err != nil {
		t.Errorf("Failed to insert timestamp:", err)
		return
	}

	_, err = db.Exec("INSERT INTO foo(id, ts) VALUES(3, ?)", "nonsense")
	if err != nil {
		t.Errorf("Failed to insert nonsense:", err)
		return
	}

	rows, err := db.Query("SELECT id, ts FROM foo ORDER BY id ASC")
	if err != nil {
		t.Errorf("Unable to query foo table:", err)
		return
	}

	seen := 0
	for rows.Next() {
		var id int
		var ts time.Time

		if err := rows.Scan(&id, &ts); err != nil {
			t.Errorf("Unable to scan results:", err)
			continue
		}

		if id == 1 {
			seen += 1
			if !timestamp1.Equal(ts) {
				t.Errorf("Value for id 1 should be %v, not %v", timestamp1, ts)
			}
		}

		if id == 2 {
			seen += 1
			if !timestamp2.Equal(ts) {
				t.Errorf("Value for id 2 should be %v, not %v", timestamp2, ts)
			}
		}
	}

	if seen != 2 {
		t.Errorf("Expected to see two valid timestamps")
	}

	// make sure "nonsense" triggered an error
	err = rows.Err()
	if err == nil || !strings.Contains(err.Error(), "cannot parse \"nonsense\"") {
		t.Errorf("Expected error from \"nonsense\" timestamp")
	}
}


func TestBoolean(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Errorf("Failed to open database:", err)
		return
	}
	
	defer os.Remove("./foo.db")

	_, err = db.Exec("CREATE TABLE foo(id INTEGER, fbool BOOLEAN)")
	if err != nil {
		t.Errorf("Failed to create table:", err)
		return
	}
	
	bool1 := true
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(1, ?)", bool1)
	if err != nil {
		t.Errorf("Failed to insert boolean:", err)
		return
	}

	bool2 := false
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(2, ?)", bool2)
	if err != nil {
		t.Errorf("Failed to insert boolean:", err)
		return
	}

	bool3 := "nonsense"
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(3, ?)", bool3)
	if err != nil {
		t.Errorf("Failed to insert nonsense:", err)
		return
	}

	rows, err := db.Query("SELECT id, fbool FROM foo where fbool is ?", bool1)
	if err != nil {
		t.Errorf("Unable to query foo table:", err)
		return
	}
	counter := 0
	
	var id int
	var fbool bool
	
	for rows.Next(){
		if err := rows.Scan(&id, &fbool); err != nil {
			t.Errorf("Unable to scan results:", err)
			return 
		}
		counter ++
	}
			
	if counter != 1{
		t.Errorf("Expected 1 row but %v", counter)
		return 	
	}
	
	if id!=1 && fbool != true {
		t.Errorf("Value for id 1 should be %v, not %v", bool1, fbool)
		return
	} 
	
				
	rows, err = db.Query("SELECT id, fbool FROM foo where fbool is ?", bool2)
	if err != nil {
		t.Errorf("Unable to query foo table:", err)
		return
	}
	
	counter = 0

	for rows.Next(){
		if err := rows.Scan(&id, &fbool); err != nil {
			t.Errorf("Unable to scan results:", err)
			return 
		}
		counter ++
	}
			
	if counter != 1{
		t.Errorf("Expected 1 row but %v", counter)
		return 	
	}
	
	if id != 2 && fbool != false {
		t.Errorf("Value for id 2 should be %v, not %v", bool2, fbool)
		return
	}
	

	// make sure "nonsense" triggered an error
	rows, err = db.Query("SELECT id, fbool FROM foo where id=?;", 3)
	if err != nil {
		t.Errorf("Unable to query foo table:", err)
		return
	}

	rows.Next()
	err = rows.Scan(&id, &fbool)
	if err == nil {
		t.Errorf("Expected error from \"nonsense\" bool")
	}
}