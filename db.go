// Copyright 2017 Canonical Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif

// Wrapper around sqlite3_db_config() for invoking the
// SQLITE_DBCONFIG_NO_CKPT_ON_CLOSE opcode, since there's no way to use C
// varargs from Go.
int databaseNoCheckpointOnClose(sqlite3 *db, int value, int *pValue) {
  return sqlite3_db_config(db, SQLITE_DBCONFIG_NO_CKPT_ON_CLOSE, value, pValue);
}
*/
import "C"

import (
	"os"
	"time"
)

// DatabaseFilename returns the path to the database file associated with the
// given connection.
func DatabaseFilename(conn *SQLiteConn) string {
	return C.GoString(C.sqlite3_db_filename(conn.db, C.CString("main")))
}

// DatabaseSize returns the size in bytes of the database file
// associated with the given connection. If the file does not exists,
// or can't be accessed, -1 is returned.
func DatabaseSize(conn *SQLiteConn) int64 {
	info, err := os.Stat(DatabaseFilename(conn))
	if err != nil {
		return -1
	}
	return info.Size()
}

// DatabaseModTime returns the modification time of the database file
// associated with the given connection. If the file does not exists,
// or can't be accessed, the zero value for Time is returned.
func DatabaseModTime(conn *SQLiteConn) time.Time {
	info, err := os.Stat(DatabaseFilename(conn))
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// DatabaseNoCheckpointOnClose disables checkpoint attempts when a
// database is closed.
func DatabaseNoCheckpointOnClose(conn *SQLiteConn) error {
	db := conn.db
	var value C.int
	if rc := C.databaseNoCheckpointOnClose(db, 1, &value); rc != 0 {
		return Error{Code: ErrNo(rc)}
	}

	// The SQLITE_DBCONFIG_NO_CKPT_ON_CLOSE opcode is supposed to save back
	// to our variable the current value of the setting. So let's double
	// check that it was actually changed to 1.
	if int(value) != 1 {
		return Error{Code: ErrInternal}
	}

	return nil
}
