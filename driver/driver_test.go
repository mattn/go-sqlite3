// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "go-sqlite3-test-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestOpen(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatalf("Failed to open database: %s", err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}

	if stat, err := os.Stat(tempFilename); err != nil || stat.IsDir() {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}

	tempFilename = TempFilename(t)
	defer os.Remove(tempFilename)

	// Open Driver Directly
	drv := &SQLiteDriver{}
	conn, err := drv.Open(tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("Failed to create connection to database")
	}
	defer conn.Close()

	stmt, err := conn.Prepare("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec([]driver.Value{}); err != nil {
		t.Fatalf("Failed to exec statement: %s", err)
	}

	// Verify database has been created
	if _, err := os.Stat(tempFilename); os.IsNotExist(err) {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}
}

func TestOpenInvalidDSN(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	// Open Invalid DSN
	drv := &SQLiteDriver{}
	conn, err := drv.Open(fmt.Sprintf("%s?%35%2%%43?test=false", tempFilename))
	if err == nil {
		t.Fatal("Connection created while error was expected")
	}
	if conn != nil {
		t.Fatal("Conection created while error was expected")
	}
}

func TestOpenConfigDSN(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	cfg := NewConfig()
	cfg.Database = tempFilename

	db, err := sql.Open("sqlite3", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("Failed to open database: %s", err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}

	if _, err := os.Stat(tempFilename); os.IsNotExist(err) {
		t.Fatalf("Failed to create database: '%s'; %s", tempFilename, err)
	}

	// Test Open Empry Database location
	cfg.Database = ""

	db, err = sql.Open("sqlite3", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err == nil {
		t.Fatalf("Table created while error was expected")
	}
}

func TestInvalidConnectHook(t *testing.T) {
	driverName := "sqlite3_invalid_connecthook"
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return fmt.Errorf("ConnectHook Error")
		},
	})

	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err == nil {
		t.Fatalf("Table created while error was expected")
	}
}

func TestInvalidExtension(t *testing.T) {
	driverName := "sqlite3_invalid_extension"
	sql.Register(driverName, &SQLiteDriver{
		Extensions: []string{
			"invalid.extension",
		},
	})

	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("create table if not exists foo (id integer)")
	if err == nil {
		t.Fatalf("Table created while error was expected")
	}

	tempFilename = TempFilename(t)
	defer os.Remove(tempFilename)

	driverName = "sqlite3_conn_invalid_extension"
	var driverConn *SQLiteConn
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			driverConn = conn
			return nil
		},
	})

	db, err = sql.Open(driverName, tempFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("SELECT 1;")
	if err != nil {
		t.Fatalf("Failed to exec ping statement")
	}

	if err := driverConn.LoadExtension("invalid.extension", ""); err == nil {
		t.Fatal("Extension loaded while error was expected")
	}
}

func TestOpenReadonly(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)

	db1, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", tempFilename))
	if err != nil {
		t.Fatal(err)
	}
	db1.Exec("CREATE TABLE test (x int, y float)")

	db2, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", tempFilename))
	if err != nil {
		t.Fatal(err)
	}

	_ = db2
	_, err = db2.Exec("INSERT INTO test VALUES (1, 3.14)")
	if err == nil {
		t.Fatal("didn't expect INSERT into read-only database to work")
	}
}

func TestRecursiveTriggers(t *testing.T) {
	cases := map[string]bool{
		"?recursive_triggers=1": true,
		"?recursive_triggers=0": false,
	}
	for option, want := range cases {
		fname := TempFilename(t)
		uri := "file:" + fname + option
		db, err := sql.Open("sqlite3", uri)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uri, err)
			continue
		}
		var enabled bool
		err = db.QueryRow("PRAGMA recursive_triggers;").Scan(&enabled)
		db.Close()
		os.Remove(fname)
		if err != nil {
			t.Errorf("query recursive_triggers for %s: %v", uri, err)
			continue
		}
		if enabled != want {
			t.Errorf("\"PRAGMA recursive_triggers;\" for %q = %t; want %t", uri, enabled, want)
			continue
		}
	}
}

func TestClose(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

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
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
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
		t.Errorf("Expected %d for fetched result, but %d:", 123, result)
	}
}

func TestUpsert(t *testing.T) {
	_, n, _ := Version()
	if !(n >= 3024000) {
		t.Skip("UPSERT requires sqlite3 => 3.24.0")
	}
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (name string primary key, counter integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	for i := 0; i < 10; i++ {
		res, err := db.Exec("insert into foo(name, counter) values('key', 1) on conflict (name) do update set counter=counter+1")
		if err != nil {
			t.Fatal("Failed to upsert record:", err)
		}
		affected, _ := res.RowsAffected()
		if affected != 1 {
			t.Fatalf("Expected %d for affected rows, but %d:", 1, affected)
		}
	}
	rows, err := db.Query("select name, counter from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
	}
	defer rows.Close()

	rows.Next()

	var resultName string
	var resultCounter int
	rows.Scan(&resultName, &resultCounter)
	if resultName != "key" {
		t.Errorf("Expected %s for fetched result, but %s:", "key", resultName)
	}
	if resultCounter != 10 {
		t.Errorf("Expected %d for fetched result, but %d:", 10, resultCounter)
	}

}

func TestUpdate(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
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
	lastID, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	if expected != lastID {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastID)
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
		t.Errorf("Expected %d for fetched result, but %d:", 234, result)
	}
}

func TestDelete(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
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
	lastID, err := res.LastInsertId()
	if err != nil {
		t.Fatal("Failed to get LastInsertId:", err)
	}
	if expected != lastID {
		t.Errorf("Expected %q for last Id, but %q:", expected, lastID)
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

func TestWAL(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
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
func TestExecer(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
       create table foo (id integer); -- one comment
       insert into foo(id) values(?);
       insert into foo(id) values(?);
       insert into foo(id) values(?); -- another comment
       `, 1, 2, 3)
	if err != nil {
		t.Error("Failed to call db.Exec:", err)
	}
}

func TestQueryer(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	create table foo (id integer);
	`)
	if err != nil {
		t.Error("Failed to call db.Query:", err)
	}

	rows, err := db.Query(`
	insert into foo(id) values(?);
	insert into foo(id) values(?);
	insert into foo(id) values(?);
	select id from foo order by id;
	`, 3, 2, 1)
	if err != nil {
		t.Error("Failed to call db.Query:", err)
	}
	defer rows.Close()
	n := 1
	if rows != nil {
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				t.Error("Failed to db.Query:", err)
			}
			if id != n {
				t.Error("Failed to db.Query: not matched results")
			}
		}
	}
}

func TestStress(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	db.Exec("CREATE TABLE foo (id int);")
	db.Exec("INSERT INTO foo VALUES(1);")
	db.Exec("INSERT INTO foo VALUES(2);")
	db.Close()

	for i := 0; i < 10000; i++ {
		db, err := sql.Open("sqlite3", tempFilename)
		if err != nil {
			t.Fatal("Failed to open database:", err)
		}

		for j := 0; j < 3; j++ {
			rows, err := db.Query("select * from foo where id=1;")
			if err != nil {
				t.Error("Failed to call db.Query:", err)
			}
			for rows.Next() {
				var i int
				if err := rows.Scan(&i); err != nil {
					t.Errorf("Scan failed: %v\n", err)
				}
			}
			if err := rows.Err(); err != nil {
				t.Errorf("Post-scan failed: %v\n", err)
			}
			rows.Close()
		}
		db.Close()
	}
}

var customFunctionOnce sync.Once

func BenchmarkCustomFunctions(b *testing.B) {
	customFunctionOnce.Do(func() {
		customAdd := func(a, b int64) int64 {
			return a + b
		}

		sql.Register("sqlite3_BenchmarkCustomFunctions", &SQLiteDriver{
			ConnectHook: func(conn *SQLiteConn) error {
				// Impure function to force sqlite to reexecute it each time.
				return conn.RegisterFunc("custom_add", customAdd, false)
			},
		})
	})

	db, err := sql.Open("sqlite3_BenchmarkCustomFunctions", ":memory:")
	if err != nil {
		b.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var i int64
		err = db.QueryRow("SELECT custom_add(1,2)").Scan(&i)
		if err != nil {
			b.Fatal("Failed to run custom add:", err)
		}
	}
}

func TestSuite(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	d, err := sql.Open("sqlite3", tempFilename+"?busy_timeout=99999")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	db = &TestDB{t, d, SQLITE, sync.Once{}}
	testing.RunTests(func(string, string) (bool, error) { return true, nil }, tests)

	if !testing.Short() {
		for _, b := range benchmarks {
			fmt.Printf("%-20s", b.Name)
			r := testing.Benchmark(b.F)
			fmt.Printf("%10d %10.0f req/s\n", r.N, float64(r.N)/r.T.Seconds())
		}
	}
	db.tearDown()
}

// Dialect is a type of dialect of databases.
type Dialect int

// Dialects for databases.
const (
	SQLITE     Dialect = iota // SQLITE mean SQLite3 dialect
	POSTGRESQL                // POSTGRESQL mean PostgreSQL dialect
	MYSQL                     // MYSQL mean MySQL dialect
)

// DB provide context for the tests
type TestDB struct {
	*testing.T
	*sql.DB
	dialect Dialect
	once    sync.Once
}

var db *TestDB

// the following tables will be created and dropped during the test
var testTables = []string{"foo", "bar", "t", "bench"}

var tests = []testing.InternalTest{
	{Name: "TestResult", F: testResult},
	{Name: "TestBlobs", F: testBlobs},
	{Name: "TestMultiBlobs", F: testMultiBlobs},
	{Name: "TestManyQueryRow", F: testManyQueryRow},
	{Name: "TestTxQuery", F: testTxQuery},
	{Name: "TestPreparedStmt", F: testPreparedStmt},
}

var benchmarks = []testing.InternalBenchmark{
	{Name: "BenchmarkExec", F: benchmarkExec},
	{Name: "BenchmarkQuery", F: benchmarkQuery},
	{Name: "BenchmarkParams", F: benchmarkParams},
	{Name: "BenchmarkStmt", F: benchmarkStmt},
	{Name: "BenchmarkRows", F: benchmarkRows},
	{Name: "BenchmarkStmtRows", F: benchmarkStmtRows},
}

func (db *TestDB) mustExec(sql string, args ...interface{}) sql.Result {
	res, err := db.Exec(sql, args...)
	if err != nil {
		db.Fatalf("Error running %q: %v", sql, err)
	}
	return res
}

func (db *TestDB) tearDown() {
	for _, tbl := range testTables {
		switch db.dialect {
		case SQLITE:
			db.mustExec("drop table if exists " + tbl)
		case MYSQL, POSTGRESQL:
			db.mustExec("drop table if exists " + tbl)
		default:
			db.Fatal("unknown dialect")
		}
	}
}

// q replaces ? parameters if needed
func (db *TestDB) q(sql string) string {
	switch db.dialect {
	case POSTGRESQL: // replace with $1, $2, ..
		qrx := regexp.MustCompile(`\?`)
		n := 0
		return qrx.ReplaceAllStringFunc(sql, func(string) string {
			n++
			return "$" + strconv.Itoa(n)
		})
	}
	return sql
}

func (db *TestDB) blobType(size int) string {
	switch db.dialect {
	case SQLITE:
		return fmt.Sprintf("blob[%d]", size)
	case POSTGRESQL:
		return "bytea"
	case MYSQL:
		return fmt.Sprintf("VARBINARY(%d)", size)
	}
	panic("unknown dialect")
}

func (db *TestDB) serialPK() string {
	switch db.dialect {
	case SQLITE:
		return "integer primary key autoincrement"
	case POSTGRESQL:
		return "serial primary key"
	case MYSQL:
		return "integer primary key auto_increment"
	}
	panic("unknown dialect")
}

func (db *TestDB) now() string {
	switch db.dialect {
	case SQLITE:
		return "datetime('now')"
	case POSTGRESQL:
		return "now()"
	case MYSQL:
		return "now()"
	}
	panic("unknown dialect")
}

func makeBench() {
	if _, err := db.Exec("create table bench (n varchar(32), i integer, d double, s varchar(32), t datetime)"); err != nil {
		panic(err)
	}
	st, err := db.Prepare("insert into bench values (?, ?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer st.Close()
	for i := 0; i < 100; i++ {
		if _, err = st.Exec(nil, i, float64(i), fmt.Sprintf("%d", i), time.Now()); err != nil {
			panic(err)
		}
	}
}

// testResult is test for result
func testResult(t *testing.T) {
	db.tearDown()
	db.mustExec("create temporary table test (id " + db.serialPK() + ", name varchar(10))")

	for i := 1; i < 3; i++ {
		r := db.mustExec(db.q("insert into test (name) values (?)"), fmt.Sprintf("row %d", i))
		n, err := r.RowsAffected()
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Errorf("got %v, want %v", n, 1)
		}
		n, err = r.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		if n != int64(i) {
			t.Errorf("got %v, want %v", n, i)
		}
	}
	if _, err := db.Exec("error!"); err == nil {
		t.Fatalf("expected error")
	}
}

// testBlobs is test for blobs
func testBlobs(t *testing.T) {
	db.tearDown()
	var blob = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	db.mustExec("create table foo (id integer primary key, bar " + db.blobType(16) + ")")
	db.mustExec(db.q("insert into foo (id, bar) values(?,?)"), 0, blob)

	want := fmt.Sprintf("%x", blob)

	b := make([]byte, 16)
	err := db.QueryRow(db.q("select bar from foo where id = ?"), 0).Scan(&b)
	got := fmt.Sprintf("%x", b)
	if err != nil {
		t.Errorf("[]byte scan: %v", err)
	} else if got != want {
		t.Errorf("for []byte, got %q; want %q", got, want)
	}

	err = db.QueryRow(db.q("select bar from foo where id = ?"), 0).Scan(&got)
	want = string(blob)
	if err != nil {
		t.Errorf("string scan: %v", err)
	} else if got != want {
		t.Errorf("for string, got %q; want %q", got, want)
	}
}

func testMultiBlobs(t *testing.T) {
	db.tearDown()
	db.mustExec("create table foo (id integer primary key, bar " + db.blobType(16) + ")")
	var blob0 = []byte{0, 1, 2, 3, 4, 5, 6, 7}
	db.mustExec(db.q("insert into foo (id, bar) values(?,?)"), 0, blob0)
	var blob1 = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	db.mustExec(db.q("insert into foo (id, bar) values(?,?)"), 1, blob1)

	r, err := db.Query(db.q("select bar from foo order by id"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if !r.Next() {
		if r.Err() != nil {
			t.Fatal(err)
		}
		t.Fatal("expected one rows")
	}

	want0 := fmt.Sprintf("%x", blob0)
	b0 := make([]byte, 8)
	err = r.Scan(&b0)
	if err != nil {
		t.Fatal(err)
	}
	got0 := fmt.Sprintf("%x", b0)

	if !r.Next() {
		if r.Err() != nil {
			t.Fatal(err)
		}
		t.Fatal("expected one rows")
	}

	want1 := fmt.Sprintf("%x", blob1)
	b1 := make([]byte, 16)
	err = r.Scan(&b1)
	if err != nil {
		t.Fatal(err)
	}
	got1 := fmt.Sprintf("%x", b1)
	if got0 != want0 {
		t.Errorf("for []byte, got %q; want %q", got0, want0)
	}
	if got1 != want1 {
		t.Errorf("for []byte, got %q; want %q", got1, want1)
	}
}

// testManyQueryRow is test for many query row
func testManyQueryRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	db.tearDown()
	db.mustExec("create table foo (id integer primary key, name varchar(50))")
	db.mustExec(db.q("insert into foo (id, name) values(?,?)"), 1, "bob")
	var name string
	for i := 0; i < 10000; i++ {
		err := db.QueryRow(db.q("select name from foo where id = ?"), 1).Scan(&name)
		if err != nil || name != "bob" {
			t.Fatalf("on query %d: err=%v, name=%q", i, err, name)
		}
	}
}

// testTxQuery is test for transactional query
func testTxQuery(t *testing.T) {
	db.tearDown()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("create table foo (id integer primary key, name varchar(50))")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec(db.q("insert into foo (id, name) values(?,?)"), 1, "bob")
	if err != nil {
		t.Fatal(err)
	}

	r, err := tx.Query(db.q("select name from foo where id = ?"), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if !r.Next() {
		if r.Err() != nil {
			t.Fatal(err)
		}
		t.Fatal("expected one rows")
	}

	var name string
	err = r.Scan(&name)
	if err != nil {
		t.Fatal(err)
	}
}

// testPreparedStmt is test for prepared statement
func testPreparedStmt(t *testing.T) {
	db.tearDown()
	db.mustExec("CREATE TABLE t (count INT)")
	sel, err := db.Prepare("SELECT count FROM t ORDER BY count DESC")
	if err != nil {
		t.Fatalf("prepare 1: %v", err)
	}
	ins, err := db.Prepare(db.q("INSERT INTO t (count) VALUES (?)"))
	if err != nil {
		t.Fatalf("prepare 2: %v", err)
	}

	for n := 1; n <= 3; n++ {
		if _, err := ins.Exec(n); err != nil {
			t.Fatalf("insert(%d) = %v", n, err)
		}
	}

	const nRuns = 10
	var wg sync.WaitGroup
	for i := 0; i < nRuns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				count := 0
				if err := sel.QueryRow().Scan(&count); err != nil && err != sql.ErrNoRows {
					t.Errorf("Query: %v", err)
					return
				}
				if _, err := ins.Exec(rand.Intn(100)); err != nil {
					t.Errorf("Insert: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// Benchmarks need to use panic() since b.Error errors are lost when
// running via testing.Benchmark() I would like to run these via go
// test -bench but calling Benchmark() from a benchmark test
// currently hangs go.

// benchmarkExec is benchmark for exec
func benchmarkExec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := db.Exec("select 1"); err != nil {
			panic(err)
		}
	}
}

// benchmarkQuery is benchmark for query
func benchmarkQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var n sql.NullString
		var i int
		var f float64
		var s string
		//		var t time.Time
		if err := db.QueryRow("select null, 1, 1.1, 'foo'").Scan(&n, &i, &f, &s); err != nil {
			panic(err)
		}
	}
}

// benchmarkParams is benchmark for params
func benchmarkParams(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var n sql.NullString
		var i int
		var f float64
		var s string
		//		var t time.Time
		if err := db.QueryRow("select ?, ?, ?, ?", nil, 1, 1.1, "foo").Scan(&n, &i, &f, &s); err != nil {
			panic(err)
		}
	}
}

// benchmarkStmt is benchmark for statement
func benchmarkStmt(b *testing.B) {
	st, err := db.Prepare("select ?, ?, ?, ?")
	if err != nil {
		panic(err)
	}
	defer st.Close()

	for n := 0; n < b.N; n++ {
		var n sql.NullString
		var i int
		var f float64
		var s string
		//		var t time.Time
		if err := st.QueryRow(nil, 1, 1.1, "foo").Scan(&n, &i, &f, &s); err != nil {
			panic(err)
		}
	}
}

// benchmarkRows is benchmark for rows
func benchmarkRows(b *testing.B) {
	db.once.Do(makeBench)

	for n := 0; n < b.N; n++ {
		var n sql.NullString
		var i int
		var f float64
		var s string
		var t time.Time
		r, err := db.Query("select * from bench")
		if err != nil {
			panic(err)
		}
		for r.Next() {
			if err = r.Scan(&n, &i, &f, &s, &t); err != nil {
				panic(err)
			}
		}
		if err = r.Err(); err != nil {
			panic(err)
		}
	}
}

// benchmarkStmtRows is benchmark for statement rows
func benchmarkStmtRows(b *testing.B) {
	db.once.Do(makeBench)

	st, err := db.Prepare("select * from bench")
	if err != nil {
		panic(err)
	}
	defer st.Close()

	for n := 0; n < b.N; n++ {
		var n sql.NullString
		var i int
		var f float64
		var s string
		var t time.Time
		r, err := st.Query()
		if err != nil {
			panic(err)
		}
		for r.Next() {
			if err = r.Scan(&n, &i, &f, &s, &t); err != nil {
				panic(err)
			}
		}
		if err = r.Err(); err != nil {
			panic(err)
		}
	}
}
