// +build sqlite_column_metadata

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_ENABLE_COLUMN_METADATA
#if defined(USE_LIBSQLITE3)
#include <sqlite3.h>
#elif defined(USE_LIBSQLCIPHER)
#include <sqlcipher/sqlite3.h>
#elif defined(USE_SQLCIPHER)
#include "sqlcipher-binding.h"
#else
#include "sqlite3-binding.h"
#endif
*/
import "C"

// ColumnTableName returns the table that is the origin of a particular result
// column in a SELECT statement.
//
// See https://www.sqlite.org/c3ref/column_database_name.html
func (s *SQLiteStmt) ColumnTableName(n int) string {
	return C.GoString(C.sqlite3_column_table_name(s.s, C.int(n)))
}
