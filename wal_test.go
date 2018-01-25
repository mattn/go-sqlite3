package sqlite3_test

import (
	"fmt"
	"strconv"
	"testing"
	"unsafe"

	"github.com/CanonicalLtd/go-sqlite3"
	"github.com/mpvl/subtest"
)

func TestWalHook_Register(t *testing.T) {
	conn, cleanup := newWalSQLiteConn()
	defer cleanup()

	arg := unsafe.Pointer(new(int))

	triggered := false
	walHook := func(hookArg unsafe.Pointer, hookConn *sqlite3.SQLiteConn, dbName string, frames int) error {
		if hookArg != arg {
			t.Errorf("wal hook invoked with unexpected arg")
		}
		if hookConn != conn {
			t.Errorf("wal hook invoked with unexpected connection")
		}
		if dbName != "main" {
			t.Errorf("wal hook invoked with unexpected database name:\n want: main\n  got: %s", dbName)
		}
		triggered = true
		return nil
	}

	sqlite3.WalHook(conn, walHook, arg)

	mustExec(conn, "CREATE TABLE test (n INT)", nil)

	if !triggered {
		t.Error("wal hook was not triggered")
	}
}

func TestWalHook_Unregister(t *testing.T) {
	conn, cleanup := newWalSQLiteConn()
	defer cleanup()

	triggered := false

	walHook := func(hookArg unsafe.Pointer, hookConn *sqlite3.SQLiteConn, dbName string, frames int) error {
		triggered = true
		return nil
	}

	sqlite3.WalHook(conn, walHook, nil)
	sqlite3.WalHook(conn, nil, nil)

	mustExec(conn, "CREATE TABLE test (n INT)", nil)

	if triggered {
		t.Error("wal hook was triggered")
	}
}

func TestWalHook_ReturningAnError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "generic error",
			err:  fmt.Errorf("boom"),
			want: sqlite3.ErrorString(sqlite3.ErrError),
		},
		{
			name: "SQLite error",
			err:  sqlite3.Error{Code: sqlite3.ErrMisuse},
			want: sqlite3.ErrorString(sqlite3.ErrMisuse),
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.name, func(t *testing.T) {
			conn, cleanup := newWalSQLiteConn()
			defer cleanup()

			walHook := func(hookArg unsafe.Pointer, hookConn *sqlite3.SQLiteConn, dbName string, frames int) error {
				return c.err // return the test case error
			}
			sqlite3.WalHook(conn, walHook, nil)

			_, err := conn.Exec("CREATE TABLE test (n INT)", nil)

			if err == nil {
				t.Fatal("hook error was not propagated")
			}
			got := err.Error()
			if got != c.want {
				t.Errorf("expected\n%q\ngot\n%q", c.want, got)
			}
		})
	}
}

func TestWalHook_NotFound(t *testing.T) {
	conn, cleanup := newWalSQLiteConn()
	defer cleanup()

	walHook := func(hookArg unsafe.Pointer, hookConn *sqlite3.SQLiteConn, dbName string, frames int) error {
		return nil
	}
	sqlite3.WalHook(conn, walHook, nil)

	// Pretend that the hook for this connection is not registered
	sqlite3.WalHookInternalDelete(conn)

	const want = "WAL hook not found"
	defer func() {
		got := recover()
		if got != want {
			t.Errorf("expected\n%q\ngot\n%q", want, got)
		}
	}()
	mustExec(conn, "CREATE TABLE foo (n INT)", nil)
}

func TestWalSize_WalFileDoesNotExists(t *testing.T) {
	conn, cleanup := newWalSQLiteConn()
	defer cleanup()

	if sqlite3.WalSize(conn) != -1 {
		t.Errorf("did not return -1, meaning missing file")
	}
}

func TestWalCheckpointV2(t *testing.T) {
	conn, cleanup := newWalSQLiteConn()
	defer cleanup()

	mustExec(conn, "CREATE TABLE test (n INT)", nil)

	// Run the checkpoint.
	size, checkpointed, err := sqlite3.WalCheckpointV2(conn, sqlite3.WalCheckpointTruncate)
	if err != nil {
		t.Fatal(err)
	}

	// Check that all frames were transferred to the database file.
	if size != 0 {
		t.Fatalf("%d frames still in the WAL", size)
	}
	if checkpointed != 0 {
		t.Fatalf("%d frames were checkpointed", checkpointed)
	}

	// Make sure the WAL got truncated.
	if sqlite3.WalSize(conn) != 0 {
		t.Fatal("WAL file size is not zero after checkpoint")
	}

	// The frames containing the table have been actually
	// checkpointed and are available in the main database file
	mustQuery(conn, "SELECT * FROM test", nil)
}

func TestWalCheckpointV2_Error(t *testing.T) {
	conn, cleanup := newWalSQLiteConn()
	defer cleanup()

	_, _, err := sqlite3.WalCheckpointV2(conn, sqlite3.WalCheckpointMode(-1))

	if err == nil {
		t.Fatal("expected error when trying to set an invalid checkpoint mode")
	}
	want := sqlite3.ErrorString(sqlite3.ErrMisuse)
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestWalAutoCheckpointPragma(t *testing.T) {
	pages := []int64{2000, 0}

	for _, p := range pages {
		subtest.Run(t, strconv.Itoa(int(p)), func(t *testing.T) {
			conn, cleanup := newFileSQLiteConn()
			defer cleanup()

			if err := sqlite3.WalAutoCheckpointPragma(conn, p); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestWalAutoCheckpointPragma_Invalid(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	err := sqlite3.WalAutoCheckpointPragma(conn, -1)

	if err == nil {
		t.Fatal("expected error when trying to set auto-checkpoint threshold to -1")
	}
	const want = "failed to set wal auto checkpoint to '-1': query 'PRAGMA wal_autocheckpoint=-1' returned '0' instead of '-1'"
	got := err.Error()
	if got != want {
		t.Errorf("expected\n%q\ngot\n%q", want, got)
	}
}

func TestShmFilename(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	got := sqlite3.ShmFilename(conn)
	want := sqlite3.DatabaseFilename(conn) + "-shm"
	if got != want {
		t.Errorf("got shm filename '%s', want '%s'", got, want)
	}
}

// Return a SQLiteConn opened against a temporary database filename
// and set to WAL journal mode.
func newWalSQLiteConn() (*sqlite3.SQLiteConn, func()) {
	conn, cleanup := newFileSQLiteConn()

	if err := sqlite3.JournalModePragma(conn, sqlite3.JournalWal); err != nil {
		panic(fmt.Sprintf("failed to set WAL journal mode: %v", err))
	}

	return conn, cleanup
}
