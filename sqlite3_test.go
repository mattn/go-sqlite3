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
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	if stat, err := os.Stat("./foo.db"); err != nil || stat.IsDir() {
		t.Error("Failed to create ./foo.db")
	}
}

func TestInsert(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	res, err := db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Fatal("Failed to insert record:", err)
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		t.Fatalf("Expected %d for affected rows, but %d:", 1, affected)
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
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
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	res, err := db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Fatal("Failed to insert record:", err)
	}
	expected, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	affected, _ := res.RowsAffected()
	if err != nil {
		t.Fatal("Failed to get RowsAffected:", err)
	}
	if affected != 1 {
		t.Fatalf("Expected %d for affected rows, but %d:", 1, affected)
	}

	res, err = db.Exec("update foo set id = 234")
	if err != nil {
		t.Fatal("Failed to update record:", err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	if expected != lastId {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastId)
	}
	affected, _ = res.RowsAffected()
	if err != nil {
		t.Fatal("Failed to get RowsAffected:", err)
	}
	if affected != 1 {
		t.Fatalf("Expected %d for affected rows, but %d:", 1, affected)
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
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
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	res, err := db.Exec("insert into foo(id) values(123)")
	if err != nil {
		t.Fatal("Failed to insert record:", err)
	}
	expected, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		t.Fatal("Failed to get RowsAffected:", err)
	}
	if affected != 1 {
		t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
	}

	res, err = db.Exec("delete from foo where id = 123")
	if err != nil {
		t.Fatal("Failed to delete record:", err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	if expected != lastId {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastId)
	}
	affected, err = res.RowsAffected()
	if err != nil {
		t.Fatal("Failed to get RowsAffected:", err)
	}
	if affected != 1 {
		t.Errorf("Expected %d for cout of affected rows, but %q:", 1, affected)
	}

	rows, err := db.Query("select id from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
	}
	defer rows.Close()

	if rows.Next() {
		t.Error("Fetched row but expected not rows")
	}
}

func TestBooleanRoundtrip(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, value BOOL)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id, value) VALUES(1, ?)", true)
	if err != nil {
		t.Fatal("Failed to insert true value:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id, value) VALUES(2, ?)", false)
	if err != nil {
		t.Fatal("Failed to insert false value:", err)
	}

	rows, err := db.Query("SELECT id, value FROM foo")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var value bool

		if err := rows.Scan(&id, &value); err != nil {
			t.Error("Unable to scan results:", err)
			continue
		}

		if id == 1 && !value {
			t.Error("Value for id 1 should be true, not false")

		} else if id == 2 && value {
			t.Error("Value for id 2 should be false, not true")
		}
	}
}

func TestTimestamp(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, ts timeSTAMP)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	timestamp1 := time.Date(2012, time.April, 6, 22, 50, 0, 0, time.UTC)
	_, err = db.Exec("INSERT INTO foo(id, ts) VALUES(1, ?)", timestamp1)
	if err != nil {
		t.Fatal("Failed to insert timestamp:", err)
	}

	timestamp2 := time.Date(2012, time.April, 6, 23, 22, 0, 0, time.UTC)
	_, err = db.Exec("INSERT INTO foo(id, ts) VALUES(2, ?)", timestamp2.Unix())
	if err != nil {
		t.Fatal("Failed to insert timestamp:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id, ts) VALUES(3, ?)", "nonsense")
	if err != nil {
		t.Fatal("Failed to insert nonsense:", err)
	}

	rows, err := db.Query("SELECT id, ts FROM foo ORDER BY id ASC")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}
	defer rows.Close()

	seen := 0
	for rows.Next() {
		var id int
		var ts time.Time

		if err := rows.Scan(&id, &ts); err != nil {
			t.Error("Unable to scan results:", err)
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
		t.Error("Expected to see two valid timestamps")
	}

	// make sure "nonsense" triggered an error
	err = rows.Err()
	if err == nil || !strings.Contains(err.Error(), "cannot parse \"nonsense\"") {
		t.Error("Expected error from \"nonsense\" timestamp")
	}
}

func TestBoolean(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer os.Remove("./foo.db")
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER, fbool BOOLEAN)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	bool1 := true
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(1, ?)", bool1)
	if err != nil {
		t.Fatal("Failed to insert boolean:", err)
	}

	bool2 := false
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(2, ?)", bool2)
	if err != nil {
		t.Fatal("Failed to insert boolean:", err)
	}

	bool3 := "nonsense"
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(3, ?)", bool3)
	if err != nil {
		t.Fatal("Failed to insert nonsense:", err)
	}

	rows, err := db.Query("SELECT id, fbool FROM foo where fbool = ?", bool1)
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}
	counter := 0

	var id int
	var fbool bool

	for rows.Next() {
		if err := rows.Scan(&id, &fbool); err != nil {
			t.Fatal("Unable to scan results:", err)
		}
		counter++
	}

	if counter != 1 {
		t.Fatalf("Expected 1 row but %v", counter)
	}

	if id != 1 && fbool != true {
		t.Fatalf("Value for id 1 should be %v, not %v", bool1, fbool)
	}

	rows, err = db.Query("SELECT id, fbool FROM foo where fbool = ?", bool2)
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	counter = 0

	for rows.Next() {
		if err := rows.Scan(&id, &fbool); err != nil {
			t.Fatal("Unable to scan results:", err)
		}
		counter++
	}

	if counter != 1 {
		t.Fatalf("Expected 1 row but %v", counter)
	}

	if id != 2 && fbool != false {
		t.Fatalf("Value for id 2 should be %v, not %v", bool2, fbool)
	}

	// make sure "nonsense" triggered an error
	rows, err = db.Query("SELECT id, fbool FROM foo where id=?;", 3)
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	rows.Next()
	err = rows.Scan(&id, &fbool)
	if err == nil {
		t.Error("Expected error from \"nonsense\" bool")
	}
}

func TestDateOnlyTimestamp(t *testing.T) {
	// make sure that timestamps without times are extractable. these are generated when
	// e.g., you use the sqlite `DATE()` built-in.

	db, er := sql.Open("sqlite3", "db.sqlite")
	if er != nil {
		t.Fatal(er)
	}
	defer func() {
		db.Close()
		os.Remove("db.sqlite")
	}()

	_, er = db.Exec(`
		CREATE TABLE test
			( entry_id INTEGER PRIMARY KEY AUTOINCREMENT
			, entry_published TIMESTAMP NOT NULL
			)
	`)
	if er != nil {
		t.Fatal(er)
	}

	_, er = db.Exec(`
		INSERT INTO test VALUES ( 1, '2012-11-04')
	`)
	if er != nil {
		t.Fatal(er)
	}

	rows, er := db.Query("SELECT entry_id, entry_published FROM test")
	if er != nil {
		t.Fatal(er)
	}
	defer rows.Close()

	if !rows.Next() {
		if er := rows.Err(); er != nil {
			t.Fatal(er)
		} else {
			t.Fatalf("Unable to extract row containing date-only timestamp")
		}
	}

	var entryId int64
	var entryPublished time.Time

	if er := rows.Scan(&entryId, &entryPublished); er != nil {
		t.Fatal(er)
	}

	if entryId != 1 {
		t.Errorf("EntryId does not match inserted value")
	}

	if entryPublished.Year() != 2012 || entryPublished.Month() != 11 || entryPublished.Day() != 4 {
		t.Errorf("Extracted time does not match inserted value")
	}
}

func TestDatetime(t *testing.T) {
	db, err := sql.Open("sqlite3", "./datetime.db")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer func() {
		db.Close()
		os.Remove("./datetime.db")
	}()

	_, err = db.Exec("DROP TABLE datetimetest")
	_, err = db.Exec(`
        CREATE TABLE datetimetest
            ( entry_id INTEGER
            , my_datetime DATETIME
            )
    `)

	if err != nil {
		t.Fatal("Failed to create table:", err)
	}
	datetime := "2006-01-02 15:04:05.003"
	_, err = db.Exec(`
        INSERT INTO datetimetest(entry_id, my_datetime) 
        VALUES(1, ?)`, datetime)
	if err != nil {
		t.Fatal("Failed to insert datetime:", err)
	}

	rows, err := db.Query(
		"SELECT entry_id, my_datetime FROM datetimetest ORDER BY entry_id ASC")
	if err != nil {
		t.Fatal("Unable to query datetimetest table:", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if er := rows.Err(); er != nil {
			t.Fatal(er)
		} else {
			t.Fatalf("Unable to extract row containing datetime")
		}
	}

	var entryId int
	var myDatetime time.Time

	if err := rows.Scan(&entryId, &myDatetime); err != nil {
		t.Error("Unable to scan results:", err)
	}

	if entryId != 1 {
		t.Errorf("EntryId does not match inserted value")
	}

	if myDatetime.Year() != 2006 ||
		myDatetime.Month() != 1 ||
		myDatetime.Day() != 2 ||
		myDatetime.Hour() != 15 ||
		myDatetime.Minute() != 4 ||
		myDatetime.Second() != 5 ||
		myDatetime.Nanosecond() != 3000000 {
		t.Errorf("Extracted time does not match inserted value")
	}

}
