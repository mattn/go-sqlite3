// Copyright (C) 2014 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.True(m.t, len(args) == 6, "six arguments expected")
	assert.Equal(m.t, "test", args[0], "module name")
	assert.Equal(m.t, "main", args[1], "db name")
	assert.Equal(m.t, "vtab", args[2], "table name")
	assert.Equal(m.t, "'1'", args[3], "first arg")
	assert.Equal(m.t, "2", args[4], "second arg")
	assert.Equal(m.t, "three", args[5], "third arg")
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

func (vc *testVTabCursor) Column(c *Context, col int) error {
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
	assert.Nil(t, err, "could not open db")
	_, err = db.Exec("CREATE VIRTUAL TABLE vtab USING test('1', 2, three)")
	assert.Nil(t, err, "could not create vtable")

	var i, value int
	rows, err := db.Query("SELECT rowid, * FROM vtab WHERE test = '3'")
	assert.Nil(t, err, "couldn't select from virtual table")
	for rows.Next() {
		rows.Scan(&i, &value)
		assert.Equal(t, intarray[i], value)
	}
	_, err = db.Exec("DROP TABLE vtab")
	assert.Nil(t, err, "couldn't drop virtual table")
}
