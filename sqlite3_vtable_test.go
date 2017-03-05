// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
// +build vtable

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

type testModule struct {
	t        *testing.T
	intarray []int
}

type testVTab struct {
	intarray []int
}

type testVTabCursor struct {
	vTab  *testVTab
	index int
}

func (m testModule) Create(c *SQLiteConn, args []string) (VTab, error) {
	if len(args) != 6 {
		m.t.Fatal("six arguments expected")
	}
	if args[0] != "test" {
		m.t.Fatal("module name")
	}
	if args[1] != "main" {
		m.t.Fatal("db name")
	}
	if args[2] != "vtab" {
		m.t.Fatal("table name")
	}
	if args[3] != "'1'" {
		m.t.Fatal("first arg")
	}
	if args[4] != "2" {
		m.t.Fatal("second arg")
	}
	if args[5] != "three" {
		m.t.Fatal("third argsecond arg")
	}
	err := c.DeclareVTab("CREATE TABLE x(test TEXT)")
	if err != nil {
		return nil, err
	}
	return &testVTab{m.intarray}, nil
}

func (m testModule) Connect(c *SQLiteConn, args []string) (VTab, error) {
	return m.Create(c, args)
}

func (m testModule) DestroyModule() {}

func (v *testVTab) BestIndex(cst []InfoConstraint, ob []InfoOrderBy) (*IndexResult, error) {
	used := make([]bool, 0, len(cst))
	for range cst {
		used = append(used, false)
	}
	return &IndexResult{
		Used:           used,
		IdxNum:         0,
		IdxStr:         "test-index",
		AlreadyOrdered: true,
		EstimatedCost:  100,
		EstimatedRows:  200,
	}, nil
}

func (v *testVTab) Disconnect() error {
	return nil
}

func (v *testVTab) Destroy() error {
	return nil
}

func (v *testVTab) Open() (VTabCursor, error) {
	return &testVTabCursor{v, 0}, nil
}

func (vc *testVTabCursor) Close() error {
	return nil
}

func (vc *testVTabCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	vc.index = 0
	return nil
}

func (vc *testVTabCursor) Next() error {
	vc.index++
	return nil
}

func (vc *testVTabCursor) EOF() bool {
	return vc.index >= len(vc.vTab.intarray)
}

func (vc *testVTabCursor) Column(c *SQLiteContext, col int) error {
	if col != 0 {
		return fmt.Errorf("column index out of bounds: %d", col)
	}
	c.ResultInt(vc.vTab.intarray[vc.index])
	return nil
}

func (vc *testVTabCursor) Rowid() (int64, error) {
	return int64(vc.index), nil
}

func TestCreateModule(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	intarray := []int{1, 2, 3}
	sql.Register("sqlite3_TestCreateModule", &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return conn.CreateModule("test", testModule{t, intarray})
		},
	})
	db, err := sql.Open("sqlite3_TestCreateModule", tempFilename)
	if err != nil {
		t.Fatalf("could not open db: %v", err)
	}
	_, err = db.Exec("CREATE VIRTUAL TABLE vtab USING test('1', 2, three)")
	if err != nil {
		t.Fatalf("could not create vtable: %v", err)
	}

	var i, value int
	rows, err := db.Query("SELECT rowid, * FROM vtab WHERE test = '3'")
	if err != nil {
		t.Fatalf("couldn't select from virtual table: %v", err)
	}
	for rows.Next() {
		rows.Scan(&i, &value)
		if intarray[i] != value {
			t.Fatalf("want %v but %v", intarray[i], value)
		}
	}
	_, err = db.Exec("DROP TABLE vtab")
	if err != nil {
		t.Fatalf("couldn't drop virtual table: %v", err)
	}
}
