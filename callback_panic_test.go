package sqlite3

import (
	"database/sql/driver"
	"errors"
	"runtime"
	"testing"
	"time"
)

// closeAfterFuncFromOtherThread keeps the calling goroutine pinned to its
// OS thread so Close runs on a different thread. A panic inside the
// registered function previously left SQLite's connection mutex held by
// the first thread, so Close would hang forever (see #1476).
func closeAfterFuncFromOtherThread(t *testing.T, fn any) {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	conn, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	c := conn.(*SQLiteConn)
	if err := c.RegisterFunc("boom", fn, false); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() { recover() }()
		rows, err := c.Query("SELECT boom()", nil)
		if err != nil {
			return
		}
		rows.Next(make([]driver.Value, 1))
	}()

	done := make(chan error, 1)
	go func() { done <- c.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not return")
	}
}

func TestRegisterFuncPanicReleasesConnectionMutex(t *testing.T) {
	closeAfterFuncFromOtherThread(t, func() int64 { panic("boom") })
}

func TestRegisterFuncErrorStillCloses(t *testing.T) {
	closeAfterFuncFromOtherThread(t, func() (int64, error) { return 0, errors.New("boom") })
}
