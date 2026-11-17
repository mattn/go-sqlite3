//go:build cgo

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestBeginTxCancelPreservesMemory(t *testing.T) {
	for _, cacheSize := range []int{0, 8} {
		for _, operation := range []string{"idle", "query", "exec"} {
			t.Run(fmt.Sprintf("cache=%d/%s", cacheSize, operation), func(t *testing.T) {
				db, err := sql.Open("sqlite3", fmt.Sprintf(":memory:?_stmt_cache_size=%d", cacheSize))
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				db.SetMaxOpenConns(1)
				db.SetMaxIdleConns(1)
				if _, err := db.Exec("CREATE TABLE session_values (value INTEGER); INSERT INTO session_values VALUES (1)"); err != nil {
					t.Fatal(err)
				}

				waitCtx, stopWaiting := context.WithTimeout(context.Background(), 5*time.Second)
				defer stopWaiting()
				var started <-chan struct{}
				var stopQuery func()
				if operation != "idle" {
					conn, err := db.Conn(waitCtx)
					if err != nil {
						t.Fatal(err)
					}
					err = conn.Raw(func(driverConn any) error {
						started, stopQuery = registerContextTestQuery(t, driverConn.(*SQLiteConn))
						return nil
					})
					conn.Close()
					if err != nil {
						t.Fatal(err)
					}
					defer stopQuery()
				}

				ctx, cancel := context.WithCancel(waitCtx)
				defer cancel()
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback()
				if _, err := tx.ExecContext(ctx, "INSERT INTO session_values VALUES (2)"); err != nil {
					t.Fatal(err)
				}

				if operation == "idle" {
					cancel()
				} else {
					done := make(chan error, 1)
					go func() {
						if operation == "query" {
							var result int64
							done <- tx.QueryRowContext(ctx, contextTestControlledQuery, contextTestMaxRows).Scan(&result)
						} else {
							_, err := tx.ExecContext(ctx, "INSERT INTO session_values "+contextTestControlledQuery, contextTestMaxRows)
							done <- err
						}
					}()
					select {
					case <-started:
					case <-waitCtx.Done():
						cancel()
						_ = stopContextTestQuery(t, stopQuery, done)
						t.Fatal("query did not start")
					}
					cancel()
					select {
					case err := <-done:
						if !errors.Is(err, context.Canceled) {
							t.Fatalf("canceled %s: got %v, want context.Canceled", operation, err)
						}
					case <-waitCtx.Done():
						_ = stopContextTestQuery(t, stopQuery, done)
						t.Fatal("query did not stop")
					}
				}

				// With one connection, the next query waits for the automatic
				// rollback. Calling tx.Rollback here could mask a lost connection.
				var count, sum int
				if err := db.QueryRowContext(waitCtx, "SELECT count(*), sum(value) FROM session_values").Scan(&count, &sum); err != nil {
					t.Fatalf("read committed data after cancellation: %v", err)
				}
				if count != 1 || sum != 1 {
					t.Fatalf("after rollback: count=%d sum=%d, want 1, 1", count, sum)
				}

				nextTx, err := db.BeginTx(waitCtx, nil)
				if err != nil {
					t.Fatalf("begin next transaction: %v", err)
				}
				defer nextTx.Rollback()
				if _, err := nextTx.ExecContext(waitCtx, "INSERT INTO session_values VALUES (3)"); err != nil {
					t.Fatal(err)
				}
				if err := nextTx.Commit(); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRowContext(waitCtx, "SELECT count(*), sum(value) FROM session_values").Scan(&count, &sum); err != nil {
					t.Fatal(err)
				}
				if count != 2 || sum != 4 {
					t.Fatalf("after next commit: count=%d sum=%d, want 2, 4", count, sum)
				}
			})
		}
	}
}

func TestTransactionCleanupFailureDiscardsConnection(t *testing.T) {
	for _, finish := range []string{"rollback", "cancel", "commit"} {
		t.Run(finish, func(t *testing.T) {
			db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "session.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
			if _, err := db.Exec("CREATE TABLE session_values (value INTEGER); INSERT INTO session_values VALUES (1)"); err != nil {
				t.Fatal(err)
			}

			waitCtx, stopWaiting := context.WithTimeout(context.Background(), 5*time.Second)
			defer stopWaiting()
			conn, err := db.Conn(waitCtx)
			if err != nil {
				t.Fatal(err)
			}
			err = conn.Raw(func(driverConn any) error {
				// Refuse transaction cleanup on this connection only. A new
				// connection can still read the committed data and begin work.
				driverConn.(*SQLiteConn).RegisterAuthorizer(func(op int, arg1, arg2, arg3 string) int {
					if op == SQLITE_TRANSACTION && (arg1 == "ROLLBACK" || finish == "commit" && arg1 == "COMMIT") {
						return SQLITE_DENY
					}
					return SQLITE_OK
				})
				return nil
			})
			conn.Close()
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(waitCtx)
			defer cancel()
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.ExecContext(ctx, "INSERT INTO session_values VALUES (2)"); err != nil {
				t.Fatal(err)
			}
			switch finish {
			case "rollback":
				if err := tx.Rollback(); err == nil {
					t.Fatal("rollback succeeded despite the callback")
				}
			case "cancel":
				cancel()
			case "commit":
				if err := tx.Commit(); err == nil {
					t.Fatal("commit succeeded despite the callback")
				}
			}

			var count, sum int
			if err := db.QueryRowContext(waitCtx, "SELECT count(*), sum(value) FROM session_values").Scan(&count, &sum); err != nil {
				t.Fatal(err)
			}
			if count != 1 || sum != 1 {
				t.Fatalf("unfinished transaction was reused: count=%d sum=%d, want 1, 1", count, sum)
			}
			nextTx, err := db.BeginTx(waitCtx, nil)
			if err != nil {
				t.Fatalf("begin next transaction: %v", err)
			}
			defer nextTx.Rollback()
			if _, err := nextTx.ExecContext(waitCtx, "INSERT INTO session_values VALUES (3)"); err != nil {
				t.Fatal(err)
			}
			if err := nextTx.Commit(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConnTransactionCleanupFailureDiscardsConnection(t *testing.T) {
	for _, finish := range []string{"rollback", "cancel", "commit"} {
		t.Run(finish, func(t *testing.T) {
			db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "session.db")+"?_foreign_keys=on")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
			if _, err := db.Exec("CREATE TABLE parent (id INTEGER PRIMARY KEY); CREATE TABLE child (id INTEGER REFERENCES parent DEFERRABLE INITIALLY DEFERRED)"); err != nil {
				t.Fatal(err)
			}

			waitCtx, stopWaiting := context.WithTimeout(context.Background(), 5*time.Second)
			defer stopWaiting()
			conn, err := db.Conn(waitCtx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if err := conn.Raw(func(driverConn any) error {
				driverConn.(*SQLiteConn).RegisterAuthorizer(func(op int, arg1, arg2, arg3 string) int {
					if op == SQLITE_TRANSACTION && arg1 == "ROLLBACK" {
						return SQLITE_DENY
					}
					return SQLITE_OK
				})
				return nil
			}); err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(waitCtx)
			defer cancel()
			tx, err := conn.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.ExecContext(ctx, "INSERT INTO child VALUES (1)"); err != nil {
				t.Fatal(err)
			}
			var finishErr error
			wantCode := ErrAuth
			switch finish {
			case "rollback":
				finishErr = tx.Rollback()
			case "cancel":
				cancel()
			case "commit":
				// The COMMIT error differs from the denied cleanup ROLLBACK,
				// so check that the original constraint error is preserved.
				finishErr = tx.Commit()
				wantCode = ErrConstraint
			}
			if finish != "cancel" {
				if !errors.Is(finishErr, driver.ErrBadConn) {
					t.Fatalf("%s: got %v, want driver.ErrBadConn", finish, finishErr)
				}
				var sqliteErr Error
				if !errors.As(finishErr, &sqliteErr) || sqliteErr.Code != wantCode {
					t.Fatalf("%s: got %v, want SQLite error %v", finish, finishErr, wantCode)
				}
			}

			// With one connection, this waits for cancellation cleanup to
			// release the pinned connection. No explicit conn.Close may mask it.
			var count int
			if err := db.QueryRowContext(waitCtx, "SELECT count(*) FROM child").Scan(&count); err != nil {
				t.Fatalf("read through replacement connection: %v", err)
			}
			if count != 0 {
				t.Fatalf("unfinished transaction was reused: count=%d, want 0", count)
			}
			if _, err := conn.ExecContext(waitCtx, "INSERT INTO parent VALUES (1)"); !errors.Is(err, sql.ErrConnDone) {
				t.Fatalf("reuse pinned connection: got %v, want sql.ErrConnDone", err)
			}
			if _, err := db.ExecContext(waitCtx, "INSERT INTO parent VALUES (1)"); err != nil {
				t.Fatalf("write through replacement connection: %v", err)
			}
		})
	}
}

func TestTransactionCleanupInAutoCommitKeepsConnection(t *testing.T) {
	for _, finish := range []string{"rollback", "commit"} {
		t.Run(finish, func(t *testing.T) {
			db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if _, err := conn.ExecContext(context.Background(), "CREATE TABLE parent (id INTEGER PRIMARY KEY); CREATE TABLE child (id INTEGER REFERENCES parent DEFERRABLE INITIALLY DEFERRED)"); err != nil {
				t.Fatal(err)
			}
			tx, err := conn.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.Exec("INSERT INTO child VALUES (1)"); err != nil {
				t.Fatal(err)
			}
			var finishErr error
			wantCode := ErrConstraint
			if finish == "rollback" {
				if _, err := tx.Exec("ROLLBACK"); err != nil {
					t.Fatal(err)
				}
				finishErr = tx.Rollback()
				wantCode = ErrError
			} else {
				finishErr = tx.Commit()
			}
			var sqliteErr Error
			if !errors.As(finishErr, &sqliteErr) || sqliteErr.Code != wantCode || errors.Is(finishErr, driver.ErrBadConn) {
				t.Fatalf("%s: got %v, want SQLite error %v without driver.ErrBadConn", finish, finishErr, wantCode)
			}
			var count int
			if err := conn.QueryRowContext(context.Background(), "SELECT count(*) FROM child").Scan(&count); err != nil || count != 0 {
				t.Fatalf("reuse connection after cleanup: count=%d err=%v, want 0, nil", count, err)
			}
			if _, err := conn.ExecContext(context.Background(), "INSERT INTO parent VALUES (1)"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSessionResetPreservesSQLState(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON; CREATE TEMP TABLE session_values (value INTEGER); BEGIN; INSERT INTO session_values VALUES (1)"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM session_values").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("transaction was reset between calls: count=%d, want 1", count)
	}
	if _, err := db.Exec("ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT count(*) FROM session_values").Scan(&count); err != nil || count != 0 {
		t.Fatalf("temporary table after rollback: count=%d err=%v, want 0, nil", count, err)
	}
	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil || foreignKeys != 1 {
		t.Fatalf("foreign_keys: got %d, %v, want 1, nil", foreignKeys, err)
	}
}
