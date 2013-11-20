package sqlite3

import (
	"./sqltest"
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TempFilename() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), "foo"+hex.EncodeToString(randBytes)+".db")
}

func TestOpen(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	if stat, err := os.Stat(tempFilename); err != nil || stat.IsDir() {
		t.Error("Failed to create ./foo.db")
	}
}

func TestClose(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (id integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	stmt, err := db.Prepare("select id from foo where id = ?")
	if err != nil {
		t.Fatal("Failed to select records:", err)
	}

	db.Close()
	_, err = stmt.Exec(1)
	if err == nil {
		t.Fatal("Failed to operate closed statement")
	}
}

func TestInsert(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
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
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
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
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
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
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
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
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer os.Remove(tempFilename)
	defer db.Close()

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, ts timeSTAMP, dt DATETIME)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	timestamp1 := time.Date(2012, time.April, 6, 22, 50, 0, 0, time.UTC)
	timestamp2 := time.Date(2006, time.January, 2, 15, 4, 5, 123456789, time.UTC)
	timestamp3 := time.Date(2012, time.November, 4, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		value    interface{}
		expected time.Time
	}{
		{"nonsense", time.Time{}},
		{"0000-00-00 00:00:00", time.Time{}},
		{timestamp1, timestamp1},
		{timestamp1.Unix(), timestamp1},
		{timestamp1.In(time.FixedZone("TEST", -7*3600)), timestamp1},
		{timestamp1.Format("2006-01-02 15:04:05.000"), timestamp1},
		{timestamp1.Format("2006-01-02T15:04:05.000"), timestamp1},
		{timestamp1.Format("2006-01-02 15:04:05"), timestamp1},
		{timestamp1.Format("2006-01-02T15:04:05"), timestamp1},
		{timestamp2, timestamp2},
		{"2006-01-02 15:04:05.123456789", timestamp2},
		{"2006-01-02T15:04:05.123456789", timestamp2},
		{"2012-11-04", timestamp3},
		{"2012-11-04 00:00", timestamp3},
		{"2012-11-04 00:00:00", timestamp3},
		{"2012-11-04 00:00:00.000", timestamp3},
		{"2012-11-04T00:00", timestamp3},
		{"2012-11-04T00:00:00", timestamp3},
		{"2012-11-04T00:00:00.000", timestamp3},
	}
	for i := range tests {
		_, err = db.Exec("INSERT INTO foo(id, ts, dt) VALUES(?, ?, ?)", i, tests[i].value, tests[i].value)
		if err != nil {
			t.Fatal("Failed to insert timestamp:", err)
		}
	}

	rows, err := db.Query("SELECT id, ts, dt FROM foo ORDER BY id ASC")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}
	defer rows.Close()

	seen := 0
	for rows.Next() {
		var id int
		var ts, dt time.Time

		if err := rows.Scan(&id, &ts, &dt); err != nil {
			t.Error("Unable to scan results:", err)
			continue
		}
		if id < 0 || id >= len(tests) {
			t.Error("Bad row id: ", id)
			continue
		}
		seen++
		if !tests[id].expected.Equal(ts) {
			t.Errorf("Timestamp value for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected, dt)
		}
		if !tests[id].expected.Equal(dt) {
			t.Errorf("Datetime value for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected, dt)
		}
	}

	if seen != len(tests) {
		t.Errorf("Expected to see %d rows", len(tests))
	}
}

func TestBoolean(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer os.Remove(tempFilename)
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

func TestFloat32(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer os.Remove(tempFilename)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id) VALUES(null)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	rows, err := db.Query("SELECT id FROM foo")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	if !rows.Next() {
		t.Fatal("Unable to query results:", err)
	}

	var id interface{}
	if err := rows.Scan(&id); err != nil {
		t.Fatal("Unable to scan results:", err)
	}
	if id != nil {
		t.Error("Expected nil but not")
	}
}

func TestNull(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer os.Remove(tempFilename)
	defer db.Close()

	rows, err := db.Query("SELECT 3.141592")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	if !rows.Next() {
		t.Fatal("Unable to query results:", err)
	}

	var v interface{}
	if err := rows.Scan(&v); err != nil {
		t.Fatal("Unable to scan results:", err)
	}
	f, ok := v.(float64)
	if !ok {
		t.Error("Expected float but not")
	}
	if f != 3.141592 {
		t.Error("Expected 3.141592 but not")
	}
}

func TestTransaction(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer os.Remove(tempFilename)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id) VALUES(1)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	rows, err := tx.Query("SELECT id from foo")
	if err != nil {
		t.Fatal("Unable to query foo table:", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatal("Failed to rollback transaction:", err)
	}

	if rows.Next() {
		t.Fatal("Unable to query results:", err)
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatal("Failed to begin transaction:", err)
	}

	_, err = tx.Exec("INSERT INTO foo(id) VALUES(1)")
	if err != nil {
		t.Fatal("Failed to insert null:", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal("Failed to commit transaction:", err)
	}

	rows, err = tx.Query("SELECT id from foo")
	if err == nil {
		t.Fatal("Expected failure to query")
	}
}

func TestWAL(t *testing.T) {
	tempFilename := TempFilename()
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	defer os.Remove(tempFilename)
	defer db.Close()
	if _, err = db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		t.Fatal("Failed to Exec PRAGMA journal_mode:", err)
	}
	if _, err = db.Exec("PRAGMA locking_mode=EXCLUSIVE;"); err != nil {
		t.Fatal("Failed to Exec PRAGMA locking_mode:", err)
	}
	if _, err = db.Exec("CREATE TABLE test (id SERIAL, user TEXT NOT NULL, name TEXT NOT NULL);"); err != nil {
		t.Fatal("Failed to Exec CREATE TABLE:", err)
	}
	if _, err = db.Exec("INSERT INTO test (user, name) VALUES ('user','name');"); err != nil {
		t.Fatal("Failed to Exec INSERT:", err)
	}

	trans, err := db.Begin()
	if err != nil {
		t.Fatal("Failed to Begin:", err)
	}
	s, err := trans.Prepare("INSERT INTO test (user, name) VALUES (?, ?);")
	if err != nil {
		t.Fatal("Failed to Prepare:", err)
	}

	var count int
	if err = trans.QueryRow("SELECT count(user) FROM test;").Scan(&count); err != nil {
		t.Fatal("Failed to QueryRow:", err)
	}
	if _, err = s.Exec("bbbb", "aaaa"); err != nil {
		t.Fatal("Failed to Exec prepared statement:", err)
	}
	if err = s.Close(); err != nil {
		t.Fatal("Failed to Close prepared statement:", err)
	}
	if err = trans.Commit(); err != nil {
		t.Fatal("Failed to Commit:", err)
	}
}

// Inspired from go-mysql driver tests
func TestDriverValues(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var typeTests = []struct {
		def  string
		sval string
		val  interface{}
	}{
		{"text", "null", nil},
		{"int", "5", int64(5)},
		{"real", "5", float64(5)},
		{"blob", "'blob'", []byte("blob")},
		{"text", "'text'", []byte("text")},
	}

	for _, tt := range typeTests {
		for _, mode := range []string{"string", "arg", "stmt"} {
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec("drop table if exists gotest"); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(fmt.Sprintf("create temporary table gotest (x %s)", tt.def)); err != nil {
				t.Fatal(err)
			}

			var r *sql.Rows

			switch mode {
			case "string":
				if _, err = tx.Exec(fmt.Sprintf("insert into gotest values (%s)", tt.sval)); err != nil {
					t.Fatal(err)
				}
				if r, err = tx.Query("select * from gotest"); err != nil {
					t.Fatal(err)
				}
			case "arg":
				if _, err = tx.Exec("insert into gotest values (?)", tt.val); err != nil {
					t.Fatal(err)
				}
				if r, err = tx.Query("select * from gotest"); err != nil {
					t.Fatal(err)
				}
			case "stmt":
				var st1, st2 *sql.Stmt
				if st1, err = tx.Prepare("insert into gotest values (?)"); err != nil {
					t.Fatal(err)
				}
				if st2, err = tx.Prepare("select * from gotest"); err != nil {
					t.Fatal(err)
				}
				if _, err = st1.Exec(tt.val); err != nil {
					t.Fatal(err)
				}
				if r, err = st2.Query(); err != nil {
					t.Fatal(err)
				}
			}

			if !r.Next() {
				if err = r.Err(); err != nil {
					t.Fatal(err)
				}
				t.Error("expected row")
			}

			var driverVal interface{}
			if err := r.Scan(&driverVal); err != nil {
				t.Fatal(err)
			}

			switch want := tt.val.(type) {
			case []byte:
				got, ok := driverVal.([]byte)
				if !ok || !bytes.Equal(got, want) {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			case nil:
				if driverVal != nil {
					t.Errorf("%v: got %v, want %v", tt, driverVal, want)
				}
			case int64:
				got, ok := driverVal.(int64)
				if !ok || !(got == want) {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			case float64:
				got, ok := driverVal.(float64)
				if !ok || !(got == want) {
					t.Errorf("%v: got %v, want %v", tt, got, want)
				}
			}

			if r.Next() {
				t.Fatal("expected exactly 1 row")
			}
			if err = r.Close(); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sqltest.RunTests(t, db, sqltest.SQLITE)
}

// TODO: Execer & Queryer currently disabled
// https://github.com/mattn/go-sqlite3/issues/82
//func TestExecer(t *testing.T) {
//	tempFilename := TempFilename()
//	db, err := sql.Open("sqlite3", tempFilename)
//	if err != nil {
//		t.Fatal("Failed to open database:", err)
//	}
//	defer os.Remove(tempFilename)
//	defer db.Close()
//
//	_, err = db.Exec(`
//	create table foo (id integer);
//	insert into foo(id) values(?);
//	insert into foo(id) values(?);
//	insert into foo(id) values(?);
//	`, 1, 2, 3)
//	if err != nil {
//		t.Error("Failed to call db.Exec:", err)
//	}
//	if err != nil {
//		t.Error("Failed to call res.RowsAffected:", err)
//	}
//}
//
//func TestQueryer(t *testing.T) {
//	tempFilename := TempFilename()
//	db, err := sql.Open("sqlite3", tempFilename)
//	if err != nil {
//		t.Fatal("Failed to open database:", err)
//	}
//	defer os.Remove(tempFilename)
//	defer db.Close()
//
//	_, err = db.Exec(`
//	create table foo (id integer);
//	`)
//	if err != nil {
//		t.Error("Failed to call db.Query:", err)
//	}
//
//	rows, err := db.Query(`
//	insert into foo(id) values(?);
//	insert into foo(id) values(?);
//	insert into foo(id) values(?);
//	select id from foo order by id;
//	`, 3, 2, 1)
//	if err != nil {
//		t.Error("Failed to call db.Query:", err)
//	}
//	defer rows.Close()
//	n := 1
//	if rows != nil {
//		for rows.Next() {
//			var id int
//			err = rows.Scan(&id)
//			if err != nil {
//				t.Error("Failed to db.Query:", err)
//			}
//			if id != n {
//				t.Error("Failed to db.Query: not matched results")
//			}
//		}
//	}
//}
