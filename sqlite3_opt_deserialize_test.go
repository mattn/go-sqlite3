//go:build cgo && (!libsqlite3 || sqlite_serialize)

package sqlite3

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"unsafe"
)

func openMemorySQLiteConn(t *testing.T) *SQLiteConn {
	t.Helper()
	conn, err := (&SQLiteDriver{}).Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return conn.(*SQLiteConn)
}

func serializeConnAfterSetup(t *testing.T, setup func(t *testing.T, c *SQLiteConn)) []byte {
	t.Helper()
	c := openMemorySQLiteConn(t)
	defer c.Close()
	if setup != nil {
		setup(t, c)
	}
	data, err := c.Serialize("")
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize: empty result")
	}
	return append([]byte(nil), data...)
}

func deserializeConn(t *testing.T, c *SQLiteConn, data []byte, schema string) {
	t.Helper()
	if err := c.Deserialize(data, schema); err != nil {
		t.Fatalf("Deserialize(%q): %v", schema, err)
	}
}

func execSQLiteConn(t *testing.T, c *SQLiteConn, query string) {
	t.Helper()
	if _, err := c.Exec(query, nil); err != nil {
		t.Fatalf("Exec(%q): %v", query, err)
	}
}

func (c *SQLiteConn) queryRowViaDriver(t *testing.T, query string, dest *int) error {
	t.Helper()
	rows, err := c.Query(query, nil)
	if err != nil {
		return err
	}
	defer rows.Close()
	vals := make([]driver.Value, 1)
	if err := rows.Next(vals); err != nil {
		return err
	}
	switch n := vals[0].(type) {
	case int64:
		*dest = int(n)
	case int:
		*dest = n
	default:
		return fmt.Errorf("unexpected type %T", vals[0])
	}
	return nil
}

func (c *SQLiteConn) queryString(t *testing.T, query string, dest *string) error {
	t.Helper()
	rows, err := c.Query(query, nil)
	if err != nil {
		return err
	}
	defer rows.Close()
	vals := make([]driver.Value, 1)
	if err := rows.Next(vals); err != nil {
		return err
	}
	s, ok := vals[0].(string)
	if !ok {
		return fmt.Errorf("unexpected type %T", vals[0])
	}
	*dest = s
	return nil
}

func (c *SQLiteConn) queryBytes(t *testing.T, query string, dest *[]byte) error {
	t.Helper()
	rows, err := c.Query(query, nil)
	if err != nil {
		return err
	}
	defer rows.Close()
	vals := make([]driver.Value, 1)
	if err := rows.Next(vals); err != nil {
		return err
	}
	b, ok := vals[0].([]byte)
	if !ok {
		return fmt.Errorf("unexpected type %T", vals[0])
	}
	*dest = append((*dest)[:0], b...)
	return nil
}

func TestDeserializeRoundTripSQLLayers(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE items (id INTEGER PRIMARY KEY, label TEXT NOT NULL)`)
		execSQLiteConn(t, c, `INSERT INTO items (label) VALUES ('alpha'), ('beta')`)
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")

	var count int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM items`, &count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count=%d, want 2", count)
	}
}

func TestDeserializePinsCallerBuffer(t *testing.T) {
	payload := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE pinned (v TEXT)`)
		execSQLiteConn(t, c, `INSERT INTO pinned (v) VALUES ('held')`)
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()

	loadAndQuery := func(buf []byte) int {
		deserializeConn(t, dst, buf, "")
		var count int
		if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM pinned`, &count); err != nil {
			t.Fatal(err)
		}
		return count
	}

	count := func() int {
		local := append([]byte(nil), payload...)
		return loadAndQuery(local)
	}()
	if count != 1 {
		t.Fatalf("count=%d, want 1", count)
	}

	dst.mu.Lock()
	pinned := dst.deserializeBuf
	dst.mu.Unlock()
	if len(pinned) == 0 {
		t.Fatal("deserializeBuf not set")
	}
	if string(pinned) != string(payload) {
		t.Fatal("pinned buffer content diverged from serialized payload")
	}
}

func TestDeserializeEmptySlice(t *testing.T) {
	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	empty := []byte{}
	if err := dst.Deserialize(empty, ""); err != nil {
		t.Fatalf("Deserialize empty: %v", err)
	}
	dst.mu.Lock()
	if dst.deserializeBuf == nil {
		t.Fatal("expected empty but non-nil pinned slice")
	}
	if len(dst.deserializeBuf) != 0 {
		t.Fatalf("pinned len=%d, want 0", len(dst.deserializeBuf))
	}
	dst.mu.Unlock()
}

func TestDeserializeInvalidPayload(t *testing.T) {
	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	garbage := []byte("not-a-sqlite-database-file")
	err := dst.Deserialize(garbage, "")
	if err != nil {
		dst.mu.Lock()
		if dst.deserializeBuf != nil {
			t.Fatal("deserializeBuf should stay unset on failure")
		}
		dst.mu.Unlock()
		return
	}
	if _, qerr := dst.Exec(`SELECT 1`, nil); qerr == nil {
		t.Fatal("expected query failure for invalid deserialized payload")
	}
}

func TestDeserializeSchemaEmptyAndMain(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE schema_probe (x INTEGER)`)
		execSQLiteConn(t, c, `INSERT INTO schema_probe (x) VALUES (7)`)
	})

	for _, schema := range []string{"", "main"} {
		schema := schema
		t.Run("schema="+strconv.Quote(schema), func(t *testing.T) {
			dst := openMemorySQLiteConn(t)
			defer dst.Close()
			deserializeConn(t, dst, data, schema)
			var v int
			if err := dst.queryRowViaDriver(t, `SELECT x FROM schema_probe`, &v); err != nil {
				t.Fatal(err)
			}
			if v != 7 {
				t.Fatalf("x=%d, want 7", v)
			}
		})
	}
}

func TestDeserializeReplacesPreviousDatabase(t *testing.T) {
	first := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE gen (n INTEGER)`)
		execSQLiteConn(t, c, `INSERT INTO gen (n) VALUES (1)`)
	})
	second := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE gen (n INTEGER)`)
		execSQLiteConn(t, c, `INSERT INTO gen (n) VALUES (1), (2), (3)`)
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, first, "")
	var count int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM gen`, &count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("after first load count=%d, want 1", count)
	}

	deserializeConn(t, dst, second, "")
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM gen`, &count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("after second load count=%d, want 3", count)
	}
}

func TestDeserializeInsertWithinBuffer(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE growable (id INTEGER PRIMARY KEY, payload BLOB)`)
	})
	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")
	execSQLiteConn(t, dst, `INSERT INTO growable (payload) VALUES (zeroblob(1024))`)
	var count int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM growable`, &count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count=%d, want 1", count)
	}
}

func TestDeserializeClosedConnection(t *testing.T) {
	data := serializeConnAfterSetup(t, nil)
	c := openMemorySQLiteConn(t)
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	err := c.Deserialize(data, "")
	if err == nil {
		t.Fatal("expected error on closed connection")
	}
	if !strings.Contains(err.Error(), "connection closed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeserializeViaDatabaseSQLConn(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE via_sql (title TEXT)`)
		execSQLiteConn(t, c, `INSERT INTO via_sql (title) VALUES ('ok')`)
	})

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sqlConn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer sqlConn.Close()

	var title string
	if err := sqlConn.Raw(func(raw any) error {
		sc := raw.(*SQLiteConn)
		if err := sc.Deserialize(data, ""); err != nil {
			return err
		}
		return sc.queryString(t, `SELECT title FROM via_sql`, &title)
	}); err != nil {
		t.Fatal(err)
	}
	if title != "ok" {
		t.Fatalf("title=%q, want ok", title)
	}
}

func TestDeserializePinnedSliceReleasedOnClose(t *testing.T) {
	data := serializeConnAfterSetup(t, nil)
	c := openMemorySQLiteConn(t)
	deserializeConn(t, c, data, "")
	c.mu.Lock()
	if len(c.deserializeBuf) == 0 {
		t.Fatal("expected pinned buffer before close")
	}
	c.mu.Unlock()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deserializeBuf != nil {
		t.Fatal("deserializeBuf should be cleared after Close")
	}
}

func TestDeserializeHugeBufferInvalidNotMallocOOM(t *testing.T) {
	if testing.Short() {
		t.Skip("allocates more than 2GiB")
	}
	if os.Getenv("GO_SQLITE3_TEST_LARGE") == "" {
		t.Skip("set GO_SQLITE3_TEST_LARGE=1 to run")
	}

	huge := sqliteMaxAllocationSize + 4096
	buf := make([]byte, huge)
	dst := openMemorySQLiteConn(t)
	defer dst.Close()

	err := dst.Deserialize(buf, "")
	if err == nil {
		t.Fatal("expected error for invalid huge payload")
	}
	if strings.Contains(strings.ToLower(err.Error()), "out of memory") {
		t.Fatalf("huge invalid buffer should not fail on sqlite3_malloc64 copy: %v", err)
	}
	dst.mu.Lock()
	if dst.deserializeBuf != nil {
		t.Fatal("deserializeBuf should stay unset on failure")
	}
	dst.mu.Unlock()
}

func TestDeserializeFileRoundTrip(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE file_rt (id INTEGER PRIMARY KEY, note TEXT)`)
		execSQLiteConn(t, c, `INSERT INTO file_rt (note) VALUES ('from-file')`)
	})

	path := TempFilename(t)
	defer os.Remove(path)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	fromFile, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open(driverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sqlConn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer sqlConn.Close()

	var note string
	if err := sqlConn.Raw(func(raw any) error {
		sc := raw.(*SQLiteConn)
		if err := sc.Deserialize(fromFile, ""); err != nil {
			return err
		}
		return sc.queryString(t, `SELECT note FROM file_rt WHERE id = 1`, &note)
	}); err != nil {
		t.Fatal(err)
	}
	if note != "from-file" {
		t.Fatalf("note=%q, want from-file", note)
	}
}

func TestDeserializeMultipleTablesAndIndexes(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE a (id INTEGER PRIMARY KEY, name TEXT)`)
		execSQLiteConn(t, c, `CREATE TABLE b (a_id INTEGER REFERENCES a(id))`)
		execSQLiteConn(t, c, `CREATE INDEX idx_a_name ON a(name)`)
		execSQLiteConn(t, c, `INSERT INTO a (name) VALUES ('x'), ('y')`)
		execSQLiteConn(t, c, `INSERT INTO b (a_id) SELECT id FROM a`)
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")

	var tables int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM sqlite_master WHERE type='table'`, &tables); err != nil {
		t.Fatal(err)
	}
	if tables != 2 {
		t.Fatalf("tables=%d, want 2", tables)
	}
	var rows int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM b`, &rows); err != nil {
		t.Fatal(err)
	}
	if rows != 2 {
		t.Fatalf("rows=%d, want 2", rows)
	}
}

func TestDeserializeAfterDeserializeDifferentBackingArrays(t *testing.T) {
	dst := openMemorySQLiteConn(t)
	defer dst.Close()

	a := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE pass (tag TEXT)`)
		execSQLiteConn(t, c, `INSERT INTO pass (tag) VALUES ('first')`)
	})
	b := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE pass (tag TEXT)`)
		execSQLiteConn(t, c, `INSERT INTO pass (tag) VALUES ('second')`)
	})

	deserializeConn(t, dst, a, "")
	var tag string
	if err := dst.queryString(t, `SELECT tag FROM pass`, &tag); err != nil {
		t.Fatal(err)
	}
	if tag != "first" {
		t.Fatalf("tag=%q, want first", tag)
	}

	deserializeConn(t, dst, b, "")
	if err := dst.queryString(t, `SELECT tag FROM pass`, &tag); err != nil {
		t.Fatal(err)
	}
	if tag != "second" {
		t.Fatalf("tag=%q, want second", tag)
	}
}

func TestDeserializeBinaryBlobRoundTrip(t *testing.T) {
	want := []byte{0, 1, 2, 255, 0, 17}
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE blobs (payload BLOB)`)
		if _, err := c.Exec(`INSERT INTO blobs (payload) VALUES (?)`, []driver.Value{want}); err != nil {
			t.Fatal(err)
		}
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")

	var got []byte
	if err := dst.queryBytes(t, `SELECT payload FROM blobs`, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("len=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte[%d]: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestDeserializeUsesProvidedBufferPointer(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE ptr_probe (n INTEGER)`)
		execSQLiteConn(t, c, `INSERT INTO ptr_probe (n) VALUES (99)`)
	})
	buf := append([]byte(nil), data...)
	dst := openMemorySQLiteConn(t)
	defer dst.Close()

	inPtr := unsafe.Pointer(unsafe.SliceData(buf))
	deserializeConn(t, dst, buf, "")

	dst.mu.Lock()
	pinned := dst.deserializeBuf
	dst.mu.Unlock()
	if uintptr(unsafe.Pointer(unsafe.SliceData(pinned))) != uintptr(inPtr) {
		t.Fatal("deserializeBuf should reference the caller-supplied backing array")
	}
}

func TestDeserializeManyRowsRoundTrip(t *testing.T) {
	const rows = 500
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE bulk (id INTEGER PRIMARY KEY, note TEXT)`)
		for i := 0; i < rows; i++ {
			_, err := c.Exec(`INSERT INTO bulk (note) VALUES (?)`, []driver.Value{fmt.Sprintf("row-%d", i)})
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")

	var count int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM bulk`, &count); err != nil {
		t.Fatal(err)
	}
	if count != rows {
		t.Fatalf("count=%d, want %d", count, rows)
	}
}

func TestDeserializePreservesViews(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE base (value INTEGER)`)
		execSQLiteConn(t, c, `INSERT INTO base (value) VALUES (5)`)
		execSQLiteConn(t, c, `CREATE VIEW base_view AS SELECT value * 2 AS doubled FROM base`)
	})

	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")

	var doubled int
	if err := dst.queryRowViaDriver(t, `SELECT doubled FROM base_view`, &doubled); err != nil {
		t.Fatal(err)
	}
	if doubled != 10 {
		t.Fatalf("doubled=%d, want 10", doubled)
	}
}

func TestDeserializeDeleteAndInsert(t *testing.T) {
	data := serializeConnAfterSetup(t, func(t *testing.T, c *SQLiteConn) {
		execSQLiteConn(t, c, `CREATE TABLE mutate (id INTEGER PRIMARY KEY)`)
		execSQLiteConn(t, c, `INSERT INTO mutate (id) VALUES (1),(2),(3)`)
	})
	dst := openMemorySQLiteConn(t)
	defer dst.Close()
	deserializeConn(t, dst, data, "")
	execSQLiteConn(t, dst, `DELETE FROM mutate WHERE id = 2`)
	execSQLiteConn(t, dst, `INSERT INTO mutate (id) VALUES (4)`)

	var count int
	if err := dst.queryRowViaDriver(t, `SELECT COUNT(*) FROM mutate`, &count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("count=%d, want 3", count)
	}
}
