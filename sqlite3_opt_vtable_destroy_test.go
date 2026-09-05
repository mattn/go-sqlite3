//go:build sqlite_vtable || vtable
// +build sqlite_vtable vtable

package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"testing"
)

type destroyErrModule struct{}

type destroyErrVTab struct{}

func (m destroyErrModule) Create(c *SQLiteConn, args []string) (VTab, error) {
	if err := c.DeclareVTab("CREATE TABLE x(test TEXT)"); err != nil {
		return nil, err
	}
	return &destroyErrVTab{}, nil
}
func (m destroyErrModule) Connect(c *SQLiteConn, args []string) (VTab, error) {
	return m.Create(c, args)
}
func (m destroyErrModule) DestroyModule() {}

func (v *destroyErrVTab) BestIndex(cst []InfoConstraint, ob []InfoOrderBy) (*IndexResult, error) {
	return &IndexResult{Used: make([]bool, len(cst)), EstimatedCost: 100}, nil
}
func (v *destroyErrVTab) Disconnect() error { return nil }
func (v *destroyErrVTab) Destroy() error    { return errors.New("boom: destroy refused") }
func (v *destroyErrVTab) Open() (VTabCursor, error) {
	return &destroyErrCursor{}, nil
}

type destroyErrCursor struct{}

func (c *destroyErrCursor) Close() error { return nil }
func (c *destroyErrCursor) Filter(idxNum int, idxStr string, vals []any) error {
	return nil
}
func (c *destroyErrCursor) Next() error                      { return nil }
func (c *destroyErrCursor) EOF() bool                        { return true }
func (c *destroyErrCursor) Column(*SQLiteContext, int) error { return nil }
func (c *destroyErrCursor) Rowid() (int64, error)            { return 0, nil }

var destroyErrDriverSeq int32

// TestVTabDestroyErrorThenClose: when xDestroy fails, SQLite retains the
// vtab object and later calls xDisconnect on it during close. The handle
// must stay registered for that call; deleting it on the error path made
// Close panic with a failed type assertion.
func TestVTabDestroyErrorThenClose(t *testing.T) {
	// Unique driver name so repeated runs (e.g. -count=2) do not panic on
	// duplicate registration.
	driverName := fmt.Sprintf("sqlite3_destroy_err_%d", atomic.AddInt32(&destroyErrDriverSeq, 1))
	sql.Register(driverName, &SQLiteDriver{
		ConnectHook: func(conn *SQLiteConn) error {
			return conn.CreateModule("boom", destroyErrModule{})
		},
	})
	db, err := sql.Open(driverName, filepath.Join(t.TempDir(), "vtab.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE VIRTUAL TABLE vtab USING boom()"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query("SELECT * FROM vtab")
	if err != nil {
		t.Fatal(err)
	}
	rows.Close()

	if _, err := db.Exec("DROP TABLE vtab"); err == nil {
		t.Fatal("DROP TABLE unexpectedly succeeded with an erroring Destroy")
	}
	// Must not panic.
	if err := db.Close(); err != nil {
		t.Logf("close err: %v", err)
	}
}
