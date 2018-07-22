// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestInvalidFunctionRegistration(t *testing.T) {
	afn := "func"
	zeroArgsFn := func(a bool) {}
	nonErrorArgsFn := func(a bool) (int, int) { return 0, 0 }

	sql.Register(fmt.Sprintf("sqlite3-%s-afn", t.Name()), &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if err := conn.RegisterFunc("afn", afn, true); err != nil {
				return err
			}

			return nil
		},
	})

	sql.Register(fmt.Sprintf("sqlite3-%s-zeroArgsFn", t.Name()), &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if err := conn.RegisterFunc("zeroArgsFn", zeroArgsFn, true); err != nil {
				return err
			}

			return nil
		},
	})

	sql.Register(fmt.Sprintf("sqlite3-%s-nonErrorArgsFn", t.Name()), &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if err := conn.RegisterFunc("nonErrorArgsFn", nonErrorArgsFn, true); err != nil {
				return err
			}

			return nil
		},
	})

	for _, s := range []string{"sqlite3-%s-afn", "sqlite3-%s-zeroArgsFn", "sqlite3-%s-nonErrorArgsFn"} {
		db, err := sql.Open(fmt.Sprintf(s, t.Name()), ":memory:")
		if err != nil {
			t.Fatal("failed to open database:", err)
		}
		defer db.Close()

		if err := db.Ping(); err == nil {
			t.Fatal("expected error from RegisterFunc")
		}
	}
}

func TestFunctionRegistration(t *testing.T) {
	addi8_16_32 := func(a int8, b int16) int32 { return int32(a) + int32(b) }
	addi64 := func(a, b int64) int64 { return a + b }
	addu8_16_32 := func(a uint8, b uint16) uint32 { return uint32(a) + uint32(b) }
	addu64 := func(a, b uint64) uint64 { return a + b }
	addiu := func(a int, b uint) int64 { return int64(a) + int64(b) }
	addf32_64 := func(a float32, b float64) float64 { return float64(a) + b }
	not := func(a bool) bool { return !a }
	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	generic := func(a interface{}) int64 {
		switch a.(type) {
		case int64:
			return 1
		case float64:
			return 2
		case []byte:
			return 3
		case string:
			return 4
		default:
			panic("unreachable")
		}
	}
	variadic := func(a, b int64, c ...int64) int64 {
		ret := a + b
		for _, d := range c {
			ret += d
		}
		return ret
	}
	variadicGeneric := func(a ...interface{}) int64 {
		return int64(len(a))
	}

	sql.Register("sqlite3_FunctionRegistration", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if err := conn.RegisterFunc("addi8_16_32", addi8_16_32, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("addi64", addi64, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("addu8_16_32", addu8_16_32, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("addu64", addu64, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("addiu", addiu, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("addf32_64", addf32_64, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("not", not, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("regex", regex, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("generic", generic, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("variadic", variadic, true); err != nil {
				return err
			}
			if err := conn.RegisterFunc("variadicGeneric", variadicGeneric, true); err != nil {
				return err
			}
			return nil
		},
	})
	db, err := sql.Open("sqlite3_FunctionRegistration", ":memory:")
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	ops := []struct {
		query    string
		expected interface{}
	}{
		{"SELECT addi8_16_32(1,2)", int32(3)},
		{"SELECT addi64(1,2)", int64(3)},
		{"SELECT addu8_16_32(1,2)", uint32(3)},
		{"SELECT addu64(1,2)", uint64(3)},
		{"SELECT addiu(1,2)", int64(3)},
		{"SELECT addf32_64(1.5,1.5)", float64(3)},
		{"SELECT not(1)", false},
		{"SELECT not(0)", true},
		{`SELECT regex("^foo.*", "foobar")`, true},
		{`SELECT regex("^foo.*", "barfoobar")`, false},
		{"SELECT generic(1)", int64(1)},
		{"SELECT generic(1.1)", int64(2)},
		{`SELECT generic(NULL)`, int64(3)},
		{`SELECT generic("foo")`, int64(4)},
		{"SELECT variadic(1,2)", int64(3)},
		{"SELECT variadic(1,2,3,4)", int64(10)},
		{"SELECT variadic(1,1,1,1,1,1,1,1,1,1)", int64(10)},
		{`SELECT variadicGeneric(1,"foo",2.3, NULL)`, int64(4)},
	}

	for _, op := range ops {
		ret := reflect.New(reflect.TypeOf(op.expected))
		err = db.QueryRow(op.query).Scan(ret.Interface())
		if err != nil {
			t.Errorf("query %q failed: %s", op.query, err)
		} else if !reflect.DeepEqual(ret.Elem().Interface(), op.expected) {
			t.Errorf("query %q returned wrong value: got %v (%T), want %v (%T)", op.query, ret.Elem().Interface(), ret.Elem().Interface(), op.expected, op.expected)
		}
	}
}

type sumAggregator int64

func (s *sumAggregator) Step(x int64) {
	*s += sumAggregator(x)
}

func (s *sumAggregator) Done() int64 {
	return int64(*s)
}

func TestAggregatorRegistration(t *testing.T) {
	customSum := func() *sumAggregator {
		var ret sumAggregator
		return &ret
	}

	sql.Register("sqlite3_AggregatorRegistration", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if err := conn.RegisterAggregator("customSum", customSum, true); err != nil {
				return err
			}
			return nil
		},
	})
	db, err := sql.Open("sqlite3_AggregatorRegistration", ":memory:")
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("create table foo (department integer, profits integer)")
	if err != nil {
		// trace feature is not implemented
		t.Skip("failed to create table:", err)
	}

	_, err = db.Exec("insert into foo values (1, 10), (1, 20), (2, 42)")
	if err != nil {
		t.Fatal("failed to insert records:", err)
	}

	tests := []struct {
		dept, sum int64
	}{
		{1, 30},
		{2, 42},
	}

	for _, test := range tests {
		var ret int64
		err = db.QueryRow("select customSum(profits) from foo where department = $1 group by department", test.dept).Scan(&ret)
		if err != nil {
			t.Fatal("query failed:", err)
		}
		if ret != test.sum {
			t.Fatalf("custom sum returned wrong value, got %d, want %d", ret, test.sum)
		}
	}
}

func rot13(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return 'A' + (r-'A'+13)%26
	case r >= 'a' && r <= 'z':
		return 'a' + (r-'a'+13)%26
	}
	return r
}

func TestCollationRegistration(t *testing.T) {
	collateRot13 := func(a, b string) int {
		ra, rb := strings.Map(rot13, a), strings.Map(rot13, b)
		return strings.Compare(ra, rb)
	}
	collateRot13Reverse := func(a, b string) int {
		return collateRot13(b, a)
	}

	sql.Register("sqlite3_CollationRegistration", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			if err := conn.RegisterCollation("rot13", collateRot13); err != nil {
				return err
			}
			if err := conn.RegisterCollation("rot13reverse", collateRot13Reverse); err != nil {
				return err
			}
			return nil
		},
	})

	db, err := sql.Open("sqlite3_CollationRegistration", ":memory:")
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	populate := []string{
		`CREATE TABLE test (s TEXT)`,
		`INSERT INTO test VALUES ("aaaa")`,
		`INSERT INTO test VALUES ("ffff")`,
		`INSERT INTO test VALUES ("qqqq")`,
		`INSERT INTO test VALUES ("tttt")`,
		`INSERT INTO test VALUES ("zzzz")`,
	}
	for _, stmt := range populate {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal("failed to populate test DB:", err)
		}
	}

	ops := []struct {
		query string
		want  []string
	}{
		{
			"SELECT * FROM test ORDER BY s COLLATE rot13 ASC",
			[]string{
				"qqqq",
				"tttt",
				"zzzz",
				"aaaa",
				"ffff",
			},
		},
		{
			"SELECT * FROM test ORDER BY s COLLATE rot13 DESC",
			[]string{
				"ffff",
				"aaaa",
				"zzzz",
				"tttt",
				"qqqq",
			},
		},
		{
			"SELECT * FROM test ORDER BY s COLLATE rot13reverse ASC",
			[]string{
				"ffff",
				"aaaa",
				"zzzz",
				"tttt",
				"qqqq",
			},
		},
		{
			"SELECT * FROM test ORDER BY s COLLATE rot13reverse DESC",
			[]string{
				"qqqq",
				"tttt",
				"zzzz",
				"aaaa",
				"ffff",
			},
		},
	}

	for _, op := range ops {
		rows, err := db.Query(op.query)
		if err != nil {
			t.Fatalf("query %q failed: %s", op.query, err)
		}
		got := []string{}
		defer rows.Close()
		for rows.Next() {
			var s string
			if err = rows.Scan(&s); err != nil {
				t.Fatalf("reading row for %q: %s", op.query, err)
			}
			got = append(got, s)
		}
		if err = rows.Err(); err != nil {
			t.Fatalf("reading rows for %q: %s", op.query, err)
		}

		if !reflect.DeepEqual(got, op.want) {
			t.Fatalf("unexpected output from %q\ngot:\n%s\n\nwant:\n%s", op.query, strings.Join(got, "\n"), strings.Join(op.want, "\n"))
		}
	}
}
