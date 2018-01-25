package sqlite3

import (
	"fmt"

	"github.com/pkg/errors"
)

// JournalMode defines all valid values for the "mode" parameter of the
// "journal_mode" PRAGMA. See https://sqlite.org/pragma.html#pragma_journal_mode.
type JournalMode string

// Available journal modes
const (
	JournalDelete   = JournalMode("delete")
	JournalTruncate = JournalMode("truncate")
	JournalPersist  = JournalMode("persist")
	JournalMemory   = JournalMode("memory")
	JournalWal      = JournalMode("wal")
	JournalOff      = JournalMode("off")
)

// JournalModePragma changes the journal mode on the given connection
// using the "journal_mode" PRAGMA.
func JournalModePragma(conn *SQLiteConn, mode JournalMode) error {
	if err := pragmaSetAndCheck(conn, "journal_mode", string(mode)); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to set journal mode to '%s'", mode))
	}
	return nil
}

// JournalSizeLimitPragma sets the maximum size of the journal.
// See https://www.sqlite.org/pragma.html#pragma_wal_autocheckpoint
func JournalSizeLimitPragma(conn *SQLiteConn, limit int64) error {
	if err := pragmaSetAndCheck(conn, "journal_size_limit", limit); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to set journal size limit to '%d'", limit))
	}
	return nil
}
