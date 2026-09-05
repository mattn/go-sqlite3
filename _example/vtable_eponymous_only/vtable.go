package main

import (
	"errors"
	"fmt"
	"strings"

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
	return &seriesTable{}, nil
}

func (m *seriesModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *seriesModule) DestroyModule() {}

type seriesTable struct{}

func (v *seriesTable) Open() (sqlite3.VTabCursor, error) {
	return &seriesCursor{}, nil
}

func (v *seriesTable) BestIndex(csts []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	// Consume equality constraints on the hidden parameter columns
	// (start, stop, step) only. A constraint on the value column must be
	// left to SQLite, otherwise it would be dropped from the WHERE clause
	// and the query would return wrong results.
	used := make([]bool, len(csts))
	var params []byte
	for c, cst := range csts {
		if !cst.Usable || cst.Op != sqlite3.OpEQ {
			continue
		}
		var p byte
		switch cst.Column {
		case 1:
			p = 's' // start
		case 2:
			p = 'e' // stop
		case 3:
			p = 't' // step
		default:
			continue
		}
		if strings.IndexByte(string(params), p) >= 0 {
			continue
		}
		used[c] = true
		params = append(params, p)
	}

	// IdxStr records which parameter each Filter argument holds, in
	// argument order.
	return &sqlite3.IndexResult{
		IdxNum: 0,
		IdxStr: string(params),
		Used:   used,
	}, nil
}

func (v *seriesTable) Disconnect() error { return nil }
func (v *seriesTable) Destroy() error    { return nil }

type seriesCursor struct {
	start int64
	stop  int64
	step  int64
	value int64
}

func (vc *seriesCursor) Column(c *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		c.ResultInt64(vc.value)
	case 1:
		c.ResultInt64(vc.start)
	case 2:
		c.ResultInt64(vc.stop)
	case 3:
		c.ResultInt64(vc.step)
	}
	return nil
}

func (vc *seriesCursor) Filter(idxNum int, idxStr string, vals []any) error {
	start, stop, step := int64(0), int64(1000), int64(1)
	for i := 0; i < len(idxStr) && i < len(vals); i++ {
		n, ok := vals[i].(int64)
		if !ok {
			return fmt.Errorf("series: argument %d must be an integer", i+1)
		}
		switch idxStr[i] {
		case 's':
			start = n
		case 'e':
			stop = n
		case 't':
			step = n
		}
	}
	if step == 0 {
		return errors.New("series: step must not be zero")
	}
	vc.start = start
	vc.stop = stop
	vc.step = step
	vc.value = start
	return nil
}

func (vc *seriesCursor) Next() error {
	vc.value += vc.step
	return nil
}

func (vc *seriesCursor) EOF() bool {
	if vc.step < 0 {
		return vc.value < vc.stop
	}
	return vc.value > vc.stop
}

func (vc *seriesCursor) Rowid() (int64, error) {
	return int64(vc.value), nil
}

func (vc *seriesCursor) Close() error {
	return nil
}
