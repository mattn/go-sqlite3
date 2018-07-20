// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo
// +build !libsqlite3

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_PREUPDATE_HOOK
#cgo LDFLAGS: -lm

#include <sqlite3-binding.h>
#include <stdlib.h>
*/
import "C"
import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

var errNilPtr = errors.New("destination pointer is nil") // embedded in descriptive error

// SQLitePreUpdateData represents all of the data available during a
// pre-update hook call.
type SQLitePreUpdateData struct {
	Conn         *SQLiteConn
	Op           int
	DatabaseName string
	TableName    string
	OldRowID     int64
	NewRowID     int64
}

// Depth returns the source path of the write, see sqlite3_preupdate_depth()
func (d *SQLitePreUpdateData) Depth() int {
	return int(C.sqlite3_preupdate_depth(d.Conn.db))
}

// Count returns the number of columns in the row
func (d *SQLitePreUpdateData) Count() int {
	return int(C.sqlite3_preupdate_count(d.Conn.db))
}

func (d *SQLitePreUpdateData) row(dest []interface{}, new bool) error {
	for i := 0; i < d.Count() && i < len(dest); i++ {
		var val *C.sqlite3_value
		var src interface{}

		// Initially I tried making this just a function pointer argument, but
		// it's absurdly complicated to pass C function pointers.
		if new {
			C.sqlite3_preupdate_new(d.Conn.db, C.int(i), &val)
		} else {
			C.sqlite3_preupdate_old(d.Conn.db, C.int(i), &val)
		}

		switch C.sqlite3_value_type(val) {
		case C.SQLITE_INTEGER:
			src = int64(C.sqlite3_value_int64(val))
		case C.SQLITE_FLOAT:
			src = float64(C.sqlite3_value_double(val))
		case C.SQLITE_BLOB:
			len := C.sqlite3_value_bytes(val)
			blobptr := C.sqlite3_value_blob(val)
			src = C.GoBytes(blobptr, len)
		case C.SQLITE_TEXT:
			len := C.sqlite3_value_bytes(val)
			cstrptr := unsafe.Pointer(C.sqlite3_value_text(val))
			src = C.GoBytes(cstrptr, len)
		case C.SQLITE_NULL:
			src = nil
		}

		err := convertAssign(&dest[i], src)
		if err != nil {
			return err
		}
	}

	return nil
}

// Old populates dest with the row data to be replaced. This works similar to
// database/sql's Rows.Scan()
func (d *SQLitePreUpdateData) Old(dest ...interface{}) error {
	if d.Op == SQLITE_INSERT {
		return errors.New("There is no old row for INSERT operations")
	}
	return d.row(dest, false)
}

// New populates dest with the replacement row data. This works similar to
// database/sql's Rows.Scan()
func (d *SQLitePreUpdateData) New(dest ...interface{}) error {
	if d.Op == SQLITE_DELETE {
		return errors.New("There is no new row for DELETE operations")
	}
	return d.row(dest, true)
}

// convertAssign copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
func convertAssign(dest, src interface{}) error {
	// Common cases, without reflect.
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = append((*d)[:0], s...)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		}
	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
			return nil
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = s.AppendFormat((*d)[:0], time.RFC3339Nano)
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		}
	}

	var sv reflect.Value

	switch d := dest.(type) {
	case *string:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = asString(src)
			return nil
		}
	case *[]byte:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes(nil, sv); ok {
			*d = b
			return nil
		}
	case *sql.RawBytes:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes([]byte(*d)[:0], sv); ok {
			*d = sql.RawBytes(b)
			return nil
		}
	case *bool:
		bv, err := driver.Bool.ConvertValue(src)
		if err == nil {
			*d = bv.(bool)
		}
		return err
	case *interface{}:
		*d = src
		return nil
	}

	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errNilPtr
	}

	if !sv.IsValid() {
		sv = reflect.ValueOf(src)
	}

	dv := reflect.Indirect(dpv)
	if sv.IsValid() && sv.Type().AssignableTo(dv.Type()) {
		switch b := src.(type) {
		case []byte:
			dv.Set(reflect.ValueOf(cloneBytes(b)))
		default:
			dv.Set(sv)
		}
		return nil
	}

	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}

	// The following conversions use a string value as an intermediate representation
	// to convert between various numeric types.
	//
	// This also allows scanning into user defined types such as "type Int int64".
	// For symmetry, also check for string destination types.
	switch dv.Kind() {
	case reflect.Ptr:
		if src == nil {
			dv.Set(reflect.Zero(dv.Type()))
			return nil
		} else {
			dv.Set(reflect.New(dv.Type().Elem()))
			return convertAssign(dv.Interface(), src)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	case reflect.String:
		switch v := src.(type) {
		case string:
			dv.SetString(v)
			return nil
		case []byte:
			dv.SetString(string(v))
			return nil
		}
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	} else {
		c := make([]byte, len(b))
		copy(c, b)
		return c
	}
}

func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

func asBytes(buf []byte, rv reflect.Value) (b []byte, ok bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10), true
	case reflect.Float32:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 32), true
	case reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64), true
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool()), true
	case reflect.String:
		s := rv.String()
		return append(buf, s...), true
	}
	return
}
