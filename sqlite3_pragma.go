package sqlite3

import (
	"database/sql/driver"
	"fmt"

	"github.com/pkg/errors"
)

// Execute a PRAGMA query that tries to set a key to a certain value, and
// checks that it actually work by comparing the value returned by the
// query and the one passed in.
func pragmaSetAndCheck(conn *SQLiteConn, key string, value interface{}) error {
	var want string

	// Convert the given value to a string and save it in the
	// 'want' variable.
	switch v := value.(type) {
	case string:
		want = v
	case int64:
		want = fmt.Sprintf("%d", v)
	default:
		panic("unsupported value type for pragma statement")
	}

	// Run the PRAGMA query
	query := fmt.Sprintf("PRAGMA %s=%s", key, want)
	rows, err := conn.Query(query, nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed query '%s'", query))
	}
	defer rows.Close()

	// Check that we have a result row
	values := make([]driver.Value, 1)
	if err := rows.Next(values); err != nil {
		return errors.Wrap(err, fmt.Sprintf("can't fetch rows for '%s'", query))
	}

	// Convert the result row to the same type as the given value
	// and check that they match.
	var got string
	switch value.(type) {
	case string:
		bytes, ok := values[0].([]byte)
		if !ok {
			return fmt.Errorf("query '%s' returned a non-byte row", query)
		}
		got = string(bytes)
	case int64:
		n, ok := values[0].(int64)
		if !ok {
			return fmt.Errorf("query '%s' returned a non-int64 row", query)
		}
		got = fmt.Sprintf("%d", n)
	}
	if got != want {
		return fmt.Errorf("query '%s' returned '%s' instead of '%s'", query, got, want)
	}
	return nil
}
