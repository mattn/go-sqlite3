// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestBooleanRoundtrip(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, value BOOL)")
	if err != nil {
		t.Fatal("failed to create table:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id, value) VALUES(1, ?)", true)
	if err != nil {
		t.Fatal("failed to insert true value:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id, value) VALUES(2, ?)", false)
	if err != nil {
		t.Fatal("failed to insert false value:", err)
	}

	rows, err := db.Query("SELECT id, value FROM foo")
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var value bool

		if err := rows.Scan(&id, &value); err != nil {
			t.Error("unable to scan results:", err)
			continue
		}

		if id == 1 && !value {
			t.Error("value for id 1 should be true, not false")

		} else if id == 2 && value {
			t.Error("value for id 2 should be false, not true")
		}
	}
}

func TestBoolean(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}

	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER, fbool BOOLEAN)")
	if err != nil {
		t.Fatal("failed to create table:", err)
	}

	bool1 := true
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(1, ?)", bool1)
	if err != nil {
		t.Fatal("failed to insert boolean:", err)
	}

	bool2 := false
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(2, ?)", bool2)
	if err != nil {
		t.Fatal("failed to insert boolean:", err)
	}

	bool3 := "nonsense"
	_, err = db.Exec("INSERT INTO foo(id, fbool) VALUES(3, ?)", bool3)
	if err != nil {
		t.Fatal("failed to insert nonsense:", err)
	}

	rows, err := db.Query("SELECT id, fbool FROM foo where fbool = ?", bool1)
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}
	counter := 0

	var id int
	var fbool bool

	for rows.Next() {
		if err := rows.Scan(&id, &fbool); err != nil {
			t.Fatal("unable to scan results:", err)
		}
		counter++
	}

	if counter != 1 {
		t.Fatalf("expected 1 row but %v", counter)
	}

	if id != 1 && !fbool {
		t.Fatalf("value for id 1 should be %v, not %v", bool1, fbool)
	}

	rows, err = db.Query("SELECT id, fbool FROM foo where fbool = ?", bool2)
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}

	counter = 0

	for rows.Next() {
		if err := rows.Scan(&id, &fbool); err != nil {
			t.Fatal("unable to scan results:", err)
		}
		counter++
	}

	if counter != 1 {
		t.Fatalf("expected 1 row but %v", counter)
	}

	if id != 2 && fbool {
		t.Fatalf("value for id 2 should be %v, not %v", bool2, fbool)
	}

	// make sure "nonsense" triggered an error
	rows, err = db.Query("SELECT id, fbool FROM foo where id=?;", 3)
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}

	rows.Next()
	err = rows.Scan(&id, &fbool)
	if err == nil {
		t.Error("expected error from \"nonsense\" bool")
	}
}

func TestFloat32(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE foo(id INTEGER)")
	if err != nil {
		t.Fatal("failed to create table:", err)
	}

	_, err = db.Exec("INSERT INTO foo(id) VALUES(null)")
	if err != nil {
		t.Fatal("failed to insert null:", err)
	}

	rows, err := db.Query("SELECT id FROM foo")
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}

	if !rows.Next() {
		t.Fatal("unable to query results:", err)
	}

	var id interface{}
	if err := rows.Scan(&id); err != nil {
		t.Fatal("unable to scan results:", err)
	}
	if id != nil {
		t.Error("expected nil but not")
	}
}

func TestNull(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT 3.141592")
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}

	if !rows.Next() {
		t.Fatal("unable to query results:", err)
	}

	var v interface{}
	if err := rows.Scan(&v); err != nil {
		t.Fatal("unable to scan results:", err)
	}
	f, ok := v.(float64)
	if !ok {
		t.Error("expected float but not")
	}
	if f != 3.141592 {
		t.Error("expected 3.141592 but not")
	}
}

func TestStringContainingZero(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec(`
	create table foo (id integer, name, extra text);
	`)
	if err != nil {
		t.Error("failed to call db.Query:", err)
	}

	const text = "foo\x00bar"

	_, err = db.Exec(`insert into foo(id, name, extra) values($1, $2, $2)`, 1, text)
	if err != nil {
		t.Error("failed to call db.Exec:", err)
	}

	row := db.QueryRow(`select id, extra from foo where id = $1 and extra = $2`, 1, text)
	if row == nil {
		t.Error("failed to call db.QueryRow")
	}

	var id int
	var extra string
	err = row.Scan(&id, &extra)
	if err != nil {
		t.Error("failed to db.Scan:", err)
	}
	if id != 1 || extra != text {
		t.Error("failed to db.QueryRow: not matched results")
	}
}

func TestDeclTypes(t *testing.T) {

	d := SQLiteDriver{}

	conn, err := d.Open(":memory:")
	if err != nil {
		t.Fatal("failed to begin transaction:", err)
	}
	defer conn.Close()

	sqlite3conn := conn.(*SQLiteConn)

	_, err = sqlite3conn.Exec("create table foo (id integer not null primary key, name text)", nil)
	if err != nil {
		t.Fatal("failed to create table:", err)
	}

	_, err = sqlite3conn.Exec("insert into foo(name) values(\"bar\")", nil)
	if err != nil {
		t.Fatal("failed to insert:", err)
	}

	rs, err := sqlite3conn.Query("select * from foo", nil)
	if err != nil {
		t.Fatal("failed to select:", err)
	}
	defer rs.Close()

	declTypes := rs.(*SQLiteRows).DeclTypes()

	if !reflect.DeepEqual(declTypes, []string{"integer", "text"}) {
		t.Fatal("unexpected declTypes:", declTypes)
	}
}

func TestNilAndEmptyBytes(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	actualNil := []byte("use this to use an actual nil not a reference to nil")
	emptyBytes := []byte{}
	for tsti, tst := range []struct {
		name          string
		columnType    string
		insertBytes   []byte
		expectedBytes []byte
	}{
		{"actual nil blob", "blob", actualNil, nil},
		{"referenced nil blob", "blob", nil, nil},
		{"empty blob", "blob", emptyBytes, emptyBytes},
		{"actual nil text", "text", actualNil, nil},
		{"referenced nil text", "text", nil, nil},
		{"empty text", "text", emptyBytes, emptyBytes},
	} {
		if _, err = db.Exec(fmt.Sprintf("create table tbl%d (txt %s)", tsti, tst.columnType)); err != nil {
			t.Fatal(tst.name, err)
		}
		if bytes.Equal(tst.insertBytes, actualNil) {
			if _, err = db.Exec(fmt.Sprintf("insert into tbl%d (txt) values (?)", tsti), nil); err != nil {
				t.Fatal(tst.name, err)
			}
		} else {
			if _, err = db.Exec(fmt.Sprintf("insert into tbl%d (txt) values (?)", tsti), &tst.insertBytes); err != nil {
				t.Fatal(tst.name, err)
			}
		}
		rows, err := db.Query(fmt.Sprintf("select txt from tbl%d", tsti))
		if err != nil {
			t.Fatal(tst.name, err)
		}
		if !rows.Next() {
			t.Fatal(tst.name, "no rows")
		}
		var scanBytes []byte
		if err = rows.Scan(&scanBytes); err != nil {
			t.Fatal(tst.name, err)
		}
		if err = rows.Err(); err != nil {
			t.Fatal(tst.name, err)
		}
		if tst.expectedBytes == nil && scanBytes != nil {
			t.Errorf("%s: %#v != %#v", tst.name, scanBytes, tst.expectedBytes)
		} else if !bytes.Equal(scanBytes, tst.expectedBytes) {
			t.Errorf("%s: %#v != %#v", tst.name, scanBytes, tst.expectedBytes)
		}
	}
}

func TestInsertNilByteSlice(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("create table blob_not_null (b blob not null)"); err != nil {
		t.Fatal(err)
	}
	var nilSlice []byte
	if _, err := db.Exec("insert into blob_not_null (b) values (?)", nilSlice); err == nil {
		t.Fatal("didn't expect INSERT to 'not null' column with a nil []byte slice to work")
	}
	zeroLenSlice := []byte{}
	if _, err := db.Exec("insert into blob_not_null (b) values (?)", zeroLenSlice); err != nil {
		t.Fatal("failed to insert zero-length slice")
	}
}
