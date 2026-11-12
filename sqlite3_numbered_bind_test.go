package sqlite3

import (
	"database/sql"
	"testing"
)

// TestNumberedBindReuseAcrossStatements covers https://github.com/mattn/go-sqlite3/issues/584:
// numbered ?NNN (and named) parameters may be reused across statements in one Exec,
// while anonymous ? parameters continue to consume distinct Go args per statement.
func TestNumberedBindReuseAcrossStatements(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE tempstorage (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		expired INTEGER NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}

	key, value, expired := "k1", "v1", int64(100)

	t.Run("anonymous", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM tempstorage`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			UPDATE tempstorage SET value = ?, expired = ? WHERE key = ?;
			INSERT INTO tempstorage(key, value, expired) SELECT ?, ?, ? WHERE changes() = 0;
		`, value, expired, key, key, value, expired)
		if err != nil {
			t.Fatal(err)
		}
		assertTempRow(t, db, key, value, expired)
	})

	t.Run("numbered", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM tempstorage`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			UPDATE tempstorage SET value = ?2, expired = ?3 WHERE key = ?1;
			INSERT INTO tempstorage(key, value, expired) SELECT ?1, ?2, ?3 WHERE changes() = 0;
		`, key, value, expired)
		if err != nil {
			t.Fatal(err)
		}
		assertTempRow(t, db, key, value, expired)
	})

	t.Run("numbered update existing", func(t *testing.T) {
		newValue, newExpired := "v2", int64(200)
		_, err := db.Exec(`
			UPDATE tempstorage SET value = ?2, expired = ?3 WHERE key = ?1;
			INSERT INTO tempstorage(key, value, expired) SELECT ?1, ?2, ?3 WHERE changes() = 0;
		`, key, newValue, newExpired)
		if err != nil {
			t.Fatal(err)
		}
		assertTempRow(t, db, key, newValue, newExpired)
	})

	t.Run("named reuse", func(t *testing.T) {
		_, err := db.Exec(`DELETE FROM tempstorage`)
		if err != nil {
			t.Fatal(err)
		}
		k, v, e := "nk", "nv", int64(300)
		_, err = db.Exec(`
			UPDATE tempstorage SET value = :value, expired = :expired WHERE key = :key;
			INSERT INTO tempstorage(key, value, expired) SELECT :key, :value, :expired WHERE changes() = 0;
		`, sql.Named("key", k), sql.Named("value", v), sql.Named("expired", e))
		if err != nil {
			t.Fatal(err)
		}
		assertTempRow(t, db, k, v, e)
	})
}

func assertTempRow(t *testing.T, db *sql.DB, key, value string, expired int64) {
	t.Helper()
	var k, v string
	var e int64
	err := db.QueryRow(`SELECT key, value, expired FROM tempstorage`).Scan(&k, &v, &e)
	if err != nil {
		t.Fatal(err)
	}
	if k != key || v != value || e != expired {
		t.Fatalf("got (%q, %q, %d), want (%q, %q, %d)", k, v, e, key, value, expired)
	}
}

func openBindReuseDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE t (a INTEGER, b INTEGER)`); err != nil {
		t.Fatal(err)
	}
	return db
}

func assertRows(t *testing.T, db *sql.DB, want [][2]int64) {
	t.Helper()
	rows, err := db.Query(`SELECT a, b FROM t ORDER BY rowid`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got [][2]int64
	for rows.Next() {
		var a, b sql.NullInt64
		if err := rows.Scan(&a, &b); err != nil {
			t.Fatal(err)
		}
		got = append(got, [2]int64{a.Int64, b.Int64})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("got rows %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got rows %v, want %v", got, want)
		}
	}
}

// A statement that only uses ?2 leaves index 1 unused; reusing it in a later
// statement must not count that gap as an anonymous parameter.
func TestNumberedBindReuseSparseIndex(t *testing.T) {
	db := openBindReuseDB(t)
	_, err := db.Exec(`
		INSERT INTO t(a) VALUES (?2);
		INSERT INTO t(a) VALUES (?2);
	`, int64(1), int64(2))
	if err != nil {
		t.Fatal(err)
	}
	assertRows(t, db, [][2]int64{{2, 0}, {2, 0}})
}

// Distinct named placeholders given positional Go args keep consuming the
// next args per statement, as before multi-statement reuse was supported.
func TestNamedPlaceholdersWithPositionalArgsAcrossStatements(t *testing.T) {
	db := openBindReuseDB(t)
	_, err := db.Exec(`
		INSERT INTO t(a) VALUES (:first);
		INSERT INTO t(a) VALUES (:second);
	`, int64(10), int64(20))
	if err != nil {
		t.Fatal(err)
	}
	assertRows(t, db, [][2]int64{{10, 0}, {20, 0}})
}

// Positional args for ":x" and "@x" stay distinct instead of both being
// renamed to "x" and bound to the last value.
func TestNamedPlaceholdersSameNameDifferentPrefixPositional(t *testing.T) {
	db := openBindReuseDB(t)
	if _, err := db.Exec(`INSERT INTO t(a, b) VALUES (:x, @x)`, int64(10), int64(20)); err != nil {
		t.Fatal(err)
	}
	assertRows(t, db, [][2]int64{{10, 20}})
}

func TestNamedBindReuseLaterDuplicateWins(t *testing.T) {
	db := openBindReuseDB(t)
	_, err := db.Exec(`
		INSERT INTO t(a) VALUES (:v);
		INSERT INTO t(a) VALUES (:v);
	`, sql.Named("v", int64(1)), sql.Named("v", int64(2)))
	if err != nil {
		t.Fatal(err)
	}
	assertRows(t, db, [][2]int64{{2, 0}, {2, 0}})
}

// Query runs the same multi-statement binding loop as Exec.
func TestNumberedAndNamedBindReuseInQuery(t *testing.T) {
	db := openBindReuseDB(t)

	var a, b int64
	err := db.QueryRow(`
		INSERT INTO t(a, b) VALUES (?1, ?2);
		SELECT ?2, ?1;
	`, int64(1), int64(2)).Scan(&a, &b)
	if err != nil {
		t.Fatal(err)
	}
	if a != 2 || b != 1 {
		t.Fatalf("numbered Query got (%d, %d), want (2, 1)", a, b)
	}

	err = db.QueryRow(`
		INSERT INTO t(a, b) VALUES (:x, :y);
		SELECT :y, :x;
	`, sql.Named("x", int64(3)), sql.Named("y", int64(4))).Scan(&a, &b)
	if err != nil {
		t.Fatal(err)
	}
	if a != 4 || b != 3 {
		t.Fatalf("named Query got (%d, %d), want (4, 3)", a, b)
	}
	assertRows(t, db, [][2]int64{{1, 2}, {3, 4}})
}
