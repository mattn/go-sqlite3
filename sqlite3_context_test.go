//go:build cgo
// +build cgo

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const contextTestTimeout = 5 * time.Second
const contextTestMaxRows = int64(1 << 60)

const contextTestControlledQuery = `
	WITH RECURSIVE numbers(value) AS (
		VALUES(0)
		UNION ALL
		SELECT value + 1 FROM numbers
		WHERE value < ? AND context_cancel_continue()
	)
	SELECT sum(value) FROM numbers`

func TestRowsContextCancelDuringStep(t *testing.T) {
	conn := openContextTestConn(t)
	started, stopQuery := registerContextTestQuery(t, conn)

	ctx, cancel := context.WithCancel(context.Background())
	// The first step runs inside QueryContext, so issue the query from a
	// goroutine and cancel while it is stepping.
	nextDone := make(chan error, 1)
	go func() {
		rows, err := conn.QueryContext(ctx, contextTestControlledQuery, contextTestQueryArgs(contextTestMaxRows))
		if err != nil {
			nextDone <- err
			return
		}
		defer rows.Close()
		nextDone <- rows.Next(make([]driver.Value, 1))
	}()

	select {
	case <-started:
	case <-time.After(contextTestTimeout):
		cancel()
		_ = stopContextTestQuery(t, stopQuery, nextDone)
		t.Fatal("query did not start")
	}
	cancel()

	select {
	case err := <-nextDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Next error = %v, want context.Canceled", err)
		}
	case <-time.After(contextTestTimeout):
		_ = stopContextTestQuery(t, stopQuery, nextDone)
		t.Fatal("Next did not return after cancellation")
	}
}

func TestRowsContextCancelBetweenRows(t *testing.T) {
	conn := openContextTestConn(t)
	ctx, cancel := context.WithCancel(context.Background())
	rows, err := conn.QueryContext(ctx, "SELECT 1 UNION ALL SELECT 2", nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	values := make([]driver.Value, 1)
	if err := rows.Next(values); err != nil {
		t.Fatalf("first Next: %v", err)
	}
	cancel()
	if err := rows.Next(values); !errors.Is(err, context.Canceled) {
		t.Fatalf("second Next error = %v, want context.Canceled", err)
	}
}

func TestRowsInterruptIgnoresInactiveRows(t *testing.T) {
	conn := openContextTestConn(t)

	idleCtx, cancelIdle := context.WithCancel(context.Background())
	defer cancelIdle()
	idleRows, err := conn.QueryContext(idleCtx, "SELECT 1 UNION ALL SELECT 2", nil)
	if err != nil {
		t.Fatalf("query idle rows: %v", err)
	}
	defer idleRows.Close()
	if err := idleRows.Next(make([]driver.Value, 1)); err != nil {
		t.Fatalf("read idle rows: %v", err)
	}

	started, stopQuery := registerContextTestQuery(t, conn)
	activeCtx, cancelActive := context.WithCancel(context.Background())
	defer cancelActive()
	nextDone := make(chan error, 1)
	activeRowsCh := make(chan *SQLiteRows, 1)
	go func() {
		activeRows, err := conn.QueryContext(activeCtx, contextTestControlledQuery, contextTestQueryArgs(contextTestMaxRows))
		if err != nil {
			nextDone <- err
			return
		}
		defer activeRows.Close()
		activeRowsCh <- activeRows.(*SQLiteRows)
		nextDone <- activeRows.Next(make([]driver.Value, 1))
	}()
	select {
	case <-started:
	case <-time.After(contextTestTimeout):
		_ = stopContextTestQuery(t, stopQuery, nextDone)
		t.Fatal("active query did not start")
	}

	// Invoke the callback guard directly so the test does not depend on scheduling.
	conn.interruptActiveRows(idleRows.(*SQLiteRows))
	if err := stopContextTestQuery(t, stopQuery, nextDone); err != nil {
		t.Fatalf("active query was interrupted: %v", err)
	}
	select {
	case rows := <-activeRowsCh:
		_ = rows
	default:
	}
}

func TestRowsLateInterruptDoesNotAffectReusedStatement(t *testing.T) {
	conn := openContextTestConn(t)
	started, stopQuery := registerContextTestQuery(t, conn)
	stmtDriver, err := conn.Prepare(contextTestControlledQuery)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	defer stmtDriver.Close()
	stmt := stmtDriver.(*SQLiteStmt)

	oldCtx, cancelOld := context.WithCancel(context.Background())
	defer cancelOld()
	oldRows, err := stmt.QueryContext(oldCtx, contextTestQueryArgs(0))
	if err != nil {
		t.Fatalf("query old rows: %v", err)
	}
	if err := oldRows.Next(make([]driver.Value, 1)); err != nil {
		t.Fatalf("read old rows: %v", err)
	}
	if err := oldRows.Close(); err != nil {
		t.Fatalf("close old rows: %v", err)
	}

	newCtx, cancelNew := context.WithCancel(context.Background())
	defer cancelNew()
	nextDone := make(chan error, 1)
	go func() {
		newRows, err := stmt.QueryContext(newCtx, contextTestQueryArgs(contextTestMaxRows))
		if err != nil {
			nextDone <- err
			return
		}
		defer newRows.Close()
		nextDone <- newRows.Next(make([]driver.Value, 1))
	}()
	select {
	case <-started:
	case <-time.After(contextTestTimeout):
		_ = stopContextTestQuery(t, stopQuery, nextDone)
		t.Fatal("replacement query did not start")
	}

	// Simulate a callback that started before oldRows was closed.
	conn.interruptActiveRows(oldRows.(*SQLiteRows))
	if err := stopContextTestQuery(t, stopQuery, nextDone); err != nil {
		t.Fatalf("replacement query was interrupted: %v", err)
	}
}

func TestRowsContextCancelAndClose(t *testing.T) {
	conn := openContextTestConn(t)
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		rows, err := conn.QueryContext(ctx, "SELECT 1 UNION ALL SELECT 2", nil)
		if err != nil {
			t.Fatalf("query cancellable rows: %v", err)
		}
		if err := rows.Next(make([]driver.Value, 1)); err != nil {
			t.Fatalf("read cancellable rows: %v", err)
		}

		start := make(chan struct{})
		cancelDone := make(chan struct{})
		closeDone := make(chan error, 1)
		go func() {
			<-start
			closeDone <- rows.Close()
		}()
		go func() {
			<-start
			cancel()
			close(cancelDone)
		}()
		close(start)
		select {
		case err := <-closeDone:
			if err != nil {
				t.Fatalf("close cancellable rows: %v", err)
			}
		case <-time.After(contextTestTimeout):
			t.Fatal("close cancellable rows timed out")
		}
		select {
		case <-cancelDone:
		case <-time.After(contextTestTimeout):
			t.Fatal("cancel cancellable rows timed out")
		}

		reusedRows, err := conn.Query("SELECT 1", nil)
		if err != nil {
			t.Fatalf("query reused connection: %v", err)
		}
		if err := reusedRows.Next(make([]driver.Value, 1)); err != nil {
			t.Fatalf("read reused connection: %v", err)
		}
		if err := reusedRows.Close(); err != nil {
			t.Fatalf("close reused rows: %v", err)
		}
	}
}

func TestRowsContextPanicClearsActiveRows(t *testing.T) {
	conn := openContextTestConn(t)
	const panicValue = "context test panic"
	if err := conn.RegisterFunc("context_cancel_panic", func() int64 {
		panic(panicValue)
	}, false); err != nil {
		t.Fatalf("register function: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The first step runs inside QueryContext, so the panic surfaces
	// there rather than at Next.
	var gotPanic any
	func() {
		defer func() {
			gotPanic = recover()
		}()
		rows, err := conn.QueryContext(ctx, "SELECT context_cancel_panic()", nil)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		defer rows.Close()
		_ = rows.Next(make([]driver.Value, 1))
	}()
	if gotPanic != panicValue {
		t.Fatalf("Next panic = %v, want %q", gotPanic, panicValue)
	}

	conn.mu.Lock()
	activeRows := conn.activeRows
	conn.mu.Unlock()
	if activeRows != nil {
		t.Fatalf("active rows = %p, want nil", activeRows)
	}
}

func TestDatabaseSQLRowsContextCancelAndReuse(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	sqlConn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("get database connection: %v", err)
	}
	defer sqlConn.Close()

	var started <-chan struct{}
	var stopQuery func()
	if err := sqlConn.Raw(func(driverConn any) error {
		sqliteConn, ok := driverConn.(*SQLiteConn)
		if !ok {
			return fmt.Errorf("driver connection type = %T, want *SQLiteConn", driverConn)
		}
		started, stopQuery = registerContextTestQuery(t, sqliteConn)
		return nil
	}); err != nil {
		t.Fatalf("access driver connection: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nextDone := make(chan error, 1)
	go func() {
		rows, err := sqlConn.QueryContext(ctx, contextTestControlledQuery, contextTestMaxRows)
		if err != nil {
			nextDone <- err
			return
		}
		if rows.Next() {
			rows.Close()
			nextDone <- errors.New("Next returned a row after cancellation")
			return
		}
		err = rows.Err()
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
		nextDone <- err
	}()
	select {
	case <-started:
	case <-time.After(contextTestTimeout):
		cancel()
		_ = stopContextTestQuery(t, stopQuery, nextDone)
		t.Fatal("query did not start")
	}
	cancel()

	select {
	case err := <-nextDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Next error = %v, want context.Canceled", err)
		}
	case <-time.After(contextTestTimeout):
		_ = stopContextTestQuery(t, stopQuery, nextDone)
		t.Fatal("Next did not return after cancellation")
	}

	var got int
	if err := sqlConn.QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil {
		t.Fatalf("query reused connection: %v", err)
	}
	if got != 1 {
		t.Fatalf("reused query = %d, want 1", got)
	}
}

func registerContextTestQuery(tb testing.TB, conn *SQLiteConn) (<-chan struct{}, func()) {
	tb.Helper()
	started := make(chan struct{})
	var startedOnce sync.Once
	var keepRunning atomic.Bool
	keepRunning.Store(true)
	if err := conn.RegisterFunc("context_cancel_continue", func() int64 {
		startedOnce.Do(func() { close(started) })
		if keepRunning.Load() {
			return 1
		}
		return 0
	}, false); err != nil {
		tb.Fatalf("register function: %v", err)
	}
	return started, func() {
		keepRunning.Store(false)
	}
}

func stopContextTestQuery(tb testing.TB, stop func(), nextDone <-chan error) error {
	tb.Helper()
	stop()
	select {
	case err := <-nextDone:
		return err
	case <-time.After(contextTestTimeout):
		tb.Fatal("controlled query did not stop")
		return nil
	}
}

func contextTestQueryArgs(maxRows int64) []driver.NamedValue {
	return []driver.NamedValue{{Ordinal: 1, Value: maxRows}}
}

func BenchmarkRowsContext(b *testing.B) {
	conn := openContextTestConn(b)
	if _, err := conn.Exec(`
		CREATE TABLE benchmark_rows(value INTEGER PRIMARY KEY);
		WITH RECURSIVE numbers(value) AS (
			VALUES(1)
			UNION ALL
			SELECT value + 1 FROM numbers WHERE value < 1000
		)
		INSERT INTO benchmark_rows SELECT value FROM numbers`, nil); err != nil {
		b.Fatalf("populate benchmark rows: %v", err)
	}
	stmtDriver, err := conn.Prepare("SELECT value FROM benchmark_rows LIMIT ?")
	if err != nil {
		b.Fatalf("prepare: %v", err)
	}
	defer stmtDriver.Close()
	stmt := stmtDriver.(*SQLiteStmt)

	contexts := []struct {
		name string
		new  func() (context.Context, context.CancelFunc)
	}{
		{
			name: "non_cancelable",
			new: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
		},
		{
			name: "cancelable",
			new: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
		},
	}

	for _, rowCount := range []int64{0, 1, 1000} {
		for _, benchmarkContext := range contexts {
			b.Run(fmt.Sprintf("rows=%d/context=%s", rowCount, benchmarkContext.name), func(b *testing.B) {
				ctx, cancel := benchmarkContext.new()
				defer cancel()
				args := []driver.NamedValue{{Ordinal: 1, Value: rowCount}}
				values := make([]driver.Value, 1)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					rows, err := stmt.QueryContext(ctx, args)
					if err != nil {
						b.Fatalf("query: %v", err)
					}
					gotRows := int64(0)
					for {
						err := rows.Next(values)
						if errors.Is(err, io.EOF) {
							break
						}
						if err != nil {
							b.Fatalf("Next: %v", err)
						}
						gotRows++
					}
					if err := rows.Close(); err != nil {
						b.Fatalf("close rows: %v", err)
					}
					if gotRows != rowCount {
						b.Fatalf("row count = %d, want %d", gotRows, rowCount)
					}
				}
				b.ReportMetric(float64(rowCount), "rows/op")
			})
		}
	}

	for _, benchmarkContext := range contexts {
		b.Run(fmt.Sprintf("close_without_next/context=%s", benchmarkContext.name), func(b *testing.B) {
			ctx, cancel := benchmarkContext.new()
			defer cancel()
			args := []driver.NamedValue{{Ordinal: 1, Value: int64(1)}}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rows, err := stmt.QueryContext(ctx, args)
				if err != nil {
					b.Fatalf("query: %v", err)
				}
				if err := rows.Close(); err != nil {
					b.Fatalf("close rows: %v", err)
				}
			}
		})
	}
}

func openContextTestConn(tb testing.TB) *SQLiteConn {
	tb.Helper()
	rawConn, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		tb.Fatalf("open SQLite connection: %v", err)
	}
	conn := rawConn.(*SQLiteConn)
	tb.Cleanup(func() {
		if err := conn.Close(); err != nil {
			tb.Errorf("close SQLite connection: %v", err)
		}
	})
	return conn
}
