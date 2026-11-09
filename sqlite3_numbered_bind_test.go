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
