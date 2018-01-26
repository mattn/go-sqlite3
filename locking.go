package sqlite3

import (
	"fmt"

	"github.com/pkg/errors"
)

// BusyTimeoutPragma changes the busy handler timeout on the given connection
// using the "busy_timeout" PRAGMA.
func BusyTimeoutPragma(conn *SQLiteConn, milliseconds int64) error {
	if err := pragmaSetAndCheck(conn, "busy_timeout", milliseconds); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to set busy timeout to '%d'", milliseconds))
	}
	return nil
}
