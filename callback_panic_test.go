//go:build cgo

package sqlite3

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
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

// ExecContext with a cancellable context runs the statement on a worker
// goroutine. A callback panic used to be re-raised there, where recover is
// impossible, killing the process instead of reaching the caller (see #1476).
func TestRegisterFuncPanicPropagatesFromExecContext(t *testing.T) {
	conn, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	c := conn.(*SQLiteConn)
	defer func() {
		if closeErr := c.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}()
	if err := c.RegisterFunc("boom", func() int64 { panic("boom") }, false); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("callback panic did not reach the caller goroutine")
		}
		if got := fmt.Sprint(r); got != "boom" {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()
	if _, execErr := c.ExecContext(ctx, "SELECT boom()", nil); execErr != nil {
		t.Fatalf("ExecContext returned an error instead of the callback panic: %v", execErr)
	}
	t.Fatal("ExecContext returned without re-raising the callback panic")
}

// The interrupt path waits for the worker before returning, so it must also
// re-raise a callback panic instead of returning the interrupt status.
func TestRegisterFuncPanicAfterContextCancel(t *testing.T) {
	conn, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	c := conn.(*SQLiteConn)
	defer func() {
		if closeErr := c.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}()
	if err := c.RegisterFunc("boom", func() int64 {
		panic("boom after cancel")
	}, false); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("callback panic did not reach the caller goroutine")
		}
		if got := fmt.Sprint(r); got != "boom after cancel" {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()
	if _, execErr := c.ExecContext(ctx, "SELECT boom()", nil); execErr != nil {
		t.Fatalf("ExecContext returned an error instead of the callback panic: %v", execErr)
	}
	t.Fatal("ExecContext returned without re-raising the callback panic")
}

// Exec without args takes the execNoArgs fast path, which must re-raise a
// callback panic itself instead of leaving it for the next statement.
func TestRegisterFuncPanicPropagatesFromExecNoArgs(t *testing.T) {
	conn, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	c := conn.(*SQLiteConn)
	defer c.Close()
	if err := c.RegisterFunc("boom", func() int64 { panic("boom") }, false); err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			if r := recover(); fmt.Sprint(r) != "boom" {
				t.Fatalf("unexpected panic value: %v", r)
			}
		}()
		_, err := c.Exec("SELECT boom()", nil)
		t.Fatalf("Exec returned without re-raising the callback panic: %v", err)
	}()

	if _, err := c.Exec("SELECT 1", nil); err != nil {
		t.Fatal(err)
	}
}
