package sqlite3

import (
	"os"
	"testing"
)

func TestWalHook(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	driver := &SQLiteDriver{}
	conn, err := driver.Open(tempFilename)
	if err != nil {
		t.Fatalf("can't open connection to %s: %v", tempFilename, err)
	}
	defer conn.Close()

	conni := conn.(*SQLiteConn)
	pragmaWAL(t, conni)

	triggered := false
	walHook := func(db string, n int) int {
		if db != "main" {
			t.Errorf("wal hook invoked with unexpected database name:\n want: main\n  got: %s", db)
		}
		triggered = true
		return 0
	}

	conni.RegisterWalHook(walHook)

	if _, err := conni.Exec("CREATE TABLE a (n INT)", nil); err != nil {
		t.Fatal("failed to execute CREATE TABLE:", err)
	}

	if !triggered {
		t.Error("wal hook was not triggered after registration")
	}

	triggered = false
	conni.RegisterWalHook(nil)

	if _, err := conni.Exec("CREATE TABLE b (n INT)", nil); err != nil {
		t.Fatal("failed to execute CREATE TABLE:", err)
	}

	if triggered {
		t.Error("wal hook was triggered after unregistration")
	}
}

func TestWalCheckpoint(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	driver := &SQLiteDriver{}
	conn, err := driver.Open(tempFilename)
	if err != nil {
		t.Fatalf("can't open connection to %s: %v", tempFilename, err)
	}
	defer conn.Close()

	conni := conn.(*SQLiteConn)
	pragmaWAL(t, conni)

	if _, err := conni.Exec("CREATE TABLE a (n INT)", nil); err != nil {
		t.Fatal("failed to execute CREATE TABLE:", err)
	}

	// Trying to use an invalid checkpoint mode results in an error
	_, _, err = conni.WalCheckpoint("main", WalCheckpointMode(-1))
	if err == nil {
		t.Error("expected error when trying to set an invalid checkpoint mode")
	}
	if err.Error() != ErrMisuse.Error() {
		t.Errorf(
			"expected invalid checkpoint mode to fail with\n%q\ngot\n%q",
			ErrMisuse.Error(), err.Error())
	}

	// Run the checkpoint.
	size, ckpt, err := conni.WalCheckpoint("main", WalCheckpointTruncate)
	if err != nil {
		t.Fatal("failed to checkpoint WAL:", err)
	}

	// Check that all frames were transferred to the database file.
	if size != 0 {
		t.Fatalf("%d frames still in the WAL", size)
	}
	if ckpt != 0 {
		t.Fatalf("only %d frames were checkpointed", ckpt)
	}
}

func pragmaWAL(t *testing.T, conn *SQLiteConn) {
	if _, err := conn.Exec("PRAGMA journal_mode=WAL;", nil); err != nil {
		t.Fatal("Failed to Exec PRAGMA journal_mode:", err)
	}
}
