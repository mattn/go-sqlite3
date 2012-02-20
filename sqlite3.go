package sqlite

/*
#include <sqlite3.h>
#include <stdlib.h>
#include <string.h>

static int
_sqlite3_bind_text(sqlite3_stmt *stmt, int n, char *p, int np) {
  return sqlite3_bind_text(stmt, n, p, np, SQLITE_TRANSIENT);
}

static int
_sqlite3_bind_blob(sqlite3_stmt *stmt, int n, void *p, int np) {
  return sqlite3_bind_blob(stmt, n, p, np, SQLITE_TRANSIENT);
}

#cgo pkg-config: sqlite3
*/
import "C"
import (
	"errors"
	"database/sql"
	"database/sql/driver"
	"unsafe"
)

func init() {
	sql.Register("sqlite3", &SQLiteDriver{})
}

type SQLiteDriver struct {
}

type SQLiteConn struct {
	db *C.sqlite3
}

type SQLiteTx struct {
	c *SQLiteConn
}

func (tx *SQLiteTx) Commit() error {
	if err := tx.c.exec("COMMIT"); err != nil {
		return err
	}
	return nil
}

func (tx *SQLiteTx) Rollback() error {
	if err := tx.c.exec("ROLLBACK"); err != nil {
		return err
	}
	return nil
}

func (c *SQLiteConn) exec(cmd string) error {
	pcmd := C.CString(cmd)
	defer C.free(unsafe.Pointer(pcmd))
	rv := C.sqlite3_exec(c.db, pcmd, nil, nil, nil)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}
	return nil
}

func (c *SQLiteConn) Begin() (driver.Tx, error) {
	if err := c.exec("BEGIN"); err != nil {
		return nil, err
	}
	return &SQLiteTx{c}, nil
}

func (d *SQLiteDriver) Open(dsn string) (driver.Conn, error) {
	if C.sqlite3_threadsafe() == 0 {
		return nil, errors.New("sqlite library was not compiled for thread-safe operation")
	}

	var db *C.sqlite3
	name := C.CString(dsn)
	defer C.free(unsafe.Pointer(name))
	rv := C.sqlite3_open_v2(name, &db,
		C.SQLITE_OPEN_FULLMUTEX|
			C.SQLITE_OPEN_READWRITE|
			C.SQLITE_OPEN_CREATE,
		nil)
	if rv != 0 {
		return nil, errors.New(C.GoString(C.sqlite3_errmsg(db)))
	}
	if db == nil {
		return nil, errors.New("sqlite succeeded without returning a database")
	}
	return &SQLiteConn{db}, nil
}

func (c *SQLiteConn) Close() error {
	s := C.sqlite3_next_stmt(c.db, nil)
	for s != nil {
		C.sqlite3_finalize(s)
		s = C.sqlite3_next_stmt(c.db, s)
	}
	rv := C.sqlite3_close(c.db)
	if rv != C.SQLITE_OK {
		return errors.New("sqlite succeeded without returning a database")
	}
	c.db = nil
	return nil
}

type SQLiteStmt struct {
	c      *SQLiteConn
	s      *C.sqlite3_stmt
	t      string
	closed bool
}

func (c *SQLiteConn) Prepare(query string) (driver.Stmt, error) {
	pquery := C.CString(query)
	defer C.free(unsafe.Pointer(pquery))
	var s *C.sqlite3_stmt
	var perror *C.char
	rv := C.sqlite3_prepare_v2(c.db, pquery, -1, &s, &perror)
	if rv != C.SQLITE_OK {
		return nil, errors.New(C.GoString(C.sqlite3_errmsg(c.db)))
	}
	var t string
	if perror != nil && C.strlen(perror) > 0 {
		t = C.GoString(perror)
	}
	return &SQLiteStmt{c: c, s: s, t: t}, nil
}

func (s *SQLiteStmt) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	rv := C.sqlite3_finalize(s.s)
	if rv != C.SQLITE_OK {
		return errors.New(C.GoString(C.sqlite3_errmsg(s.c.db)))
	}
	return nil
}

func (s *SQLiteStmt) NumInput() int {
	return int(C.sqlite3_bind_parameter_count(s.s))
}

func (s *SQLiteStmt) bind(args []driver.Value) error {
	rv := C.sqlite3_reset(s.s)
	if rv != C.SQLITE_ROW && rv != C.SQLITE_OK && rv != C.SQLITE_DONE {
		return errors.New(C.GoString(C.sqlite3_errmsg(s.c.db)))
	}

	for i, v := range args {
		n := C.int(i + 1)
		switch v := v.(type) {
		case nil:
			rv = C.sqlite3_bind_null(s.s, n)
		case string:
			if len(v) == 0 {
				b := []byte{0}
				rv = C._sqlite3_bind_text(s.s, n, (*C.char)(unsafe.Pointer(&b[0])), C.int(0))
			} else {
				b := []byte(v)
				rv = C._sqlite3_bind_text(s.s, n, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)))
			}
		case int:
			rv = C.sqlite3_bind_int(s.s, n, C.int(v))
		case int64:
			rv = C.sqlite3_bind_int64(s.s, n, C.sqlite3_int64(v))
		case byte:
			rv = C.sqlite3_bind_int(s.s, n, C.int(v))
		case bool:
			if bool(v) {
				rv = C.sqlite3_bind_int(s.s, n, -1)
			} else {
				rv = C.sqlite3_bind_int(s.s, n, 0)
			}
		case float32:
			rv = C.sqlite3_bind_double(s.s, n, C.double(v))
		case float64:
			rv = C.sqlite3_bind_double(s.s, n, C.double(v))
		case []byte:
			var p *byte
			if len(v) > 0 {
				p = &v[0]
			}
			rv = C._sqlite3_bind_blob(s.s, n, unsafe.Pointer(p), C.int(len(v)))
		}
		if rv != C.SQLITE_OK {
			return errors.New(C.GoString(C.sqlite3_errmsg(s.c.db)))
		}
	}
	return nil
}

func (s *SQLiteStmt) Query(args []driver.Value) (driver.Rows, error) {
	if err := s.bind(args); err != nil {
		return nil, err
	}
	return &SQLiteRows{s, int(C.sqlite3_column_count(s.s)), nil}, nil
}

type SQLiteResult struct {
	s *SQLiteStmt
}

func (r *SQLiteResult) LastInsertId() (int64, error) {
	return int64(C.sqlite3_last_insert_rowid(r.s.c.db)), nil
}

func (r *SQLiteResult) RowsAffected() (int64, error) {
	return int64(C.sqlite3_changes(r.s.c.db)), nil
}

func (s *SQLiteStmt) Exec(args []driver.Value) (driver.Result, error) {
	if err := s.bind(args); err != nil {
		return nil, err
	}
	rv := C.sqlite3_step(s.s)
	if rv != C.SQLITE_ROW && rv != C.SQLITE_OK && rv != C.SQLITE_DONE {
		return nil, errors.New(C.GoString(C.sqlite3_errmsg(s.c.db)))
	}
	return &SQLiteResult{s}, nil
}

type SQLiteRows struct {
	s    *SQLiteStmt
	nc   int
	cols []string
}

func (rc *SQLiteRows) Close() error {
	return rc.s.Close()
}

func (rc *SQLiteRows) Columns() []string {
	if rc.nc != len(rc.cols) {
		rc.cols = make([]string, rc.nc)
		for i := 0; i < rc.nc; i++ {
			rc.cols[i] = C.GoString(C.sqlite3_column_name(rc.s.s, C.int(i)))
		}
	}
	return rc.cols
}

func (rc *SQLiteRows) Next(dest []driver.Value) error {
	rv := C.sqlite3_step(rc.s.s)
	if rv != C.SQLITE_ROW {
		return errors.New(C.GoString(C.sqlite3_errmsg(rc.s.c.db)))
	}
	for i := range dest {
		switch C.sqlite3_column_type(rc.s.s, C.int(i)) {
		case C.SQLITE_INTEGER:
			dest[i] = int64(C.sqlite3_column_int64(rc.s.s, C.int(i)))
		case C.SQLITE_FLOAT:
			dest[i] = float64(C.sqlite3_column_double(rc.s.s, C.int(i)))
		case C.SQLITE_BLOB:
			n := int(C.sqlite3_column_bytes(rc.s.s, C.int(i)))
			p := C.sqlite3_column_blob(rc.s.s, C.int(i))
			dest[i] = (*[1 << 30]byte)(unsafe.Pointer(p))[0:n]
		case C.SQLITE_NULL:
			dest[i] = nil
		case C.SQLITE_TEXT:
			dest[i] = C.GoString((*C.char)(unsafe.Pointer(C.sqlite3_column_text(rc.s.s, C.int(i)))))
		}
	}
	return nil
}
