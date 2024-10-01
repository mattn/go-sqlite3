package main

import (
	"database/sql"
	"flag"
	"time"

	"github.com/mattn/go-sqlite3"
)

// A busy handler can be as simple as a hardcoded sleep
// We use the count along with the sleep duration to decide
// roughly when we've hit the timeout
func simpleBusyCallback(count int) int {
	const timeout = 5 * time.Second
	const delay = 1 * time.Millisecond

	if delay*time.Duration(count) >= timeout {
		// Trigger a SQLITE_BUSY error
		return 0
	}

	time.Sleep(delay)

	// Attempt to access the database again
	return 1
}

// This is a copy of the sqliteDefaultBusyCallback implementation
// from the SQLite3 source code with minor changes
//
// This is equivalent to the function that's used when the
// busy_timeout pragma is set
func defaultBusyCallback(count int) int {
	// All durations are in milliseconds
	const timeout = 5000
	delays := [...]int{1, 2, 5, 10, 15, 20, 25, 25, 25, 50, 50, 100}
	totals := [...]int{0, 1, 3, 8, 18, 33, 53, 78, 103, 128, 178, 228}

	var delay, prior int
	if count < len(delays) {
		delay = delays[count]
		prior = totals[count]
	} else {
		delay = delays[len(delays)-1]
		prior = totals[len(totals)-1] + delay*(count-len(totals))
	}

	if prior+delay > timeout {
		delay = timeout - prior

		if delay <= 0 {
			// Trigger a SQLITE_BUSY error
			return 0
		}
	}

	time.Sleep(time.Duration(delay) * time.Millisecond)

	// Attempt to access the database again
	return 1
}

func main() {
	var simple bool
	flag.BoolVar(&simple, "simple", false, "Use the simple busy handler")
	flag.Parse()

	sql.Register("sqlite3_with_busy_handler_example", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if simple {
				conn.RegisterBusyHandler(simpleBusyCallback)
			} else {
				conn.RegisterBusyHandler(defaultBusyCallback)
			}

			return nil
		},
	})
}
