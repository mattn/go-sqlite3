package sqlite3_test

import (
	"strconv"
	"testing"

	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/mpvl/subtest"
)

func TestJournalModePragma(t *testing.T) {
	modes := []sqlite3.JournalMode{
		sqlite3.JournalDelete,
		sqlite3.JournalTruncate,
		sqlite3.JournalPersist,
		sqlite3.JournalMemory,
		sqlite3.JournalWal,
		sqlite3.JournalOff,
	}
	for _, mode := range modes {
		subtest.Run(t, string(mode), func(t *testing.T) {
			conn, cleanup := newFileSQLiteConn()
			defer cleanup()

			if err := sqlite3.JournalModePragma(conn, mode); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestJournalModePragma_PragmaStatementExecutionFails(t *testing.T) {
	conn := newMemorySQLiteConn()

	// Close the connection to trigger a query error
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}

	err := sqlite3.JournalModePragma(conn, sqlite3.JournalWal)
	if err == nil {
		t.Fatal("no error returned")
	}
	want := "failed to set journal mode to 'wal': failed query 'PRAGMA journal_mode=wal': out of memory"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestJournalModePragma_Invalid(t *testing.T) {
	conn := newMemorySQLiteConn()

	err := sqlite3.JournalModePragma(conn, sqlite3.JournalMode("foo"))
	if err == nil {
		t.Fatal("expected error when trying to set an invalid journal mode")
	}
	want := "failed to set journal mode to 'foo': query 'PRAGMA journal_mode=foo' returned 'memory' instead of 'foo'"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestJournalSizeLimitPragma(t *testing.T) {
	limits := []int64{-1, 0, 1, 100, 10000}
	for _, limit := range limits {
		subtest.Run(t, strconv.Itoa(int(limit)), func(t *testing.T) {
			conn := newMemorySQLiteConn()
			if err := sqlite3.JournalSizeLimitPragma(conn, limit); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestJournalSizeLimitPragma_Invalid(t *testing.T) {
	conn := newMemorySQLiteConn()

	err := sqlite3.JournalSizeLimitPragma(conn, -1000)
	if err == nil {
		t.Fatal("expected error when setting journal size limit to -1000")
	}
	want := "failed to set journal size limit to '-1000': query 'PRAGMA journal_size_limit=-1000' returned '-1' instead of '-1000'"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}

}
