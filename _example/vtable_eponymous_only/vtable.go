package main

import (
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type seriesModule struct{}

func (m *seriesModule) EponymousOnlyModule() {}

func (m *seriesModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			value INT,
			start HIDDEN,
			stop HIDDEN,
			step HIDDEN
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &seriesTable{0, 0, 1}, nil
}

func (m *seriesModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *seriesModule) DestroyModule() {}

type seriesTable struct {
	start int64
	stop  int64
	step  int64
}

func (v *seriesTable) Open() (sqlite3.VTabCursor, error) {
	return &seriesCursor{v, 0}, nil
}

func (v *seriesTable) BestIndex(csts []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	used := make([]bool, len(csts))
	for c, cst := range csts {
		if cst.Usable && cst.Op == sqlite3.OpEQ {
			used[c] = true
		}
	}

	return &sqlite3.IndexResult{
		IdxNum: 0,
		IdxStr: "default",
		Used:   used,
	}, nil
}

func (v *seriesTable) Disconnect() error { return nil }
func (v *seriesTable) Destroy() error    { return nil }

type seriesCursor struct {
	*seriesTable
	value int64
}

func (vc *seriesCursor) Column(c *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		c.ResultInt64(vc.value)
	case 1:
		c.ResultInt64(vc.seriesTable.start)
	case 2:
		c.ResultInt64(vc.seriesTable.stop)
	case 3:
		c.ResultInt64(vc.seriesTable.step)
	}
	return nil
}

func (vc *seriesCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	switch {
	case len(vals) < 1:
		vc.seriesTable.start = 0
		vc.seriesTable.stop = 1000
		vc.value = vc.seriesTable.start
	case len(vals) < 2:
		vc.seriesTable.start = vals[0].(int64)
		vc.seriesTable.stop = 1000
		vc.value = vc.seriesTable.start
	case len(vals) < 3:
		vc.seriesTable.start = vals[0].(int64)
		vc.seriesTable.stop = vals[1].(int64)
		vc.value = vc.seriesTable.start
	case len(vals) < 4:
		vc.seriesTable.start = vals[0].(int64)
		vc.seriesTable.stop = vals[1].(int64)
		vc.seriesTable.step = vals[2].(int64)
	}

	return nil
}

func (vc *seriesCursor) Next() error {
	vc.value += vc.step
	return nil
}

func (vc *seriesCursor) EOF() bool {
	return vc.value > vc.stop
}

func (vc *seriesCursor) Rowid() (int64, error) {
	return int64(vc.value), nil
}

func (vc *seriesCursor) Close() error {
	return nil
}
