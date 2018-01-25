package sqlite3_test

import (
	"testing"

	"github.com/CanonicalLtd/go-sqlite3"
)

func TestBusyTimeoutPragma_Error(t *testing.T) {
	conn := newMemorySQLiteConn()

	err := sqlite3.BusyTimeoutPragma(conn, -1)
	if err == nil {
		t.Fatal("no failure even if passed negative milliseconds")
	}
	const want = "failed to set busy timeout to '-1': query 'PRAGMA busy_timeout=-1' returned '0' instead of '-1'"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", got, want)
	}
}
