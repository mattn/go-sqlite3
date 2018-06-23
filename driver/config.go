// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

/*
#ifndef USE_LIBSQLITE3
#include <sqlite3-binding.h>
#else
#include <sqlite3.h>
#endif
*/
import "C"
import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// SQLITE_OPEN_MUTEX_NO will force the database connection opens
	// in the multi-thread threading mode as long as the
	// single-thread mode has not been set at compile-time or start-time.
	SQLITE_OPEN_MUTEX_NO = C.SQLITE_OPEN_NOMUTEX

	// SQLITE_OPEN_MUTEX_FULL will force the database connection opens
	// in the serialized threading mode unless single-thread
	// was previously selected at compile-time or start-time.
	SQLITE_OPEN_MUTEX_FULL = C.SQLITE_OPEN_FULLMUTEX
)

// TxLock defines the Transaction Lock Behaviour.
type TxLock string

func (tx TxLock) String() string {
	return string(tx)
}

const (
	// TxLockDeferred deferred transaction behaviour. (Default)
	// Deferred means that no locks are acquired on the database
	// until the database is first accessed.
	// Thus with a deferred transaction,
	// the BEGIN statement itself does nothing to the filesystem.
	// Locks are not acquired until the first read or write operation.
	// The first read operation against a database creates a SHARED lock
	// and the first write operation creates a RESERVED lock.
	// Because the acquisition of locks is deferred until they are needed,
	// it is possible that another thread or process could create a separate transaction
	// and write to the database after the BEGIN on the current thread has executed.
	TxLockDeferred = TxLock("BEGIN")

	// TxLockImmediate immediate transaction behaviour.
	// If the transaction is immediate,
	// then RESERVED locks are acquired on all databases
	// as soon as the BEGIN command is executed,
	// without waiting for the database to be used.
	// After a BEGIN IMMEDIATE, no other database connection
	// will be able to write to the database or do a BEGIN IMMEDIATE or BEGIN EXCLUSIVE.
	// Other processes can continue to read from the database however.
	TxLockImmediate = TxLock("BEGIN IMMEDIATE")

	// TxLockExclusive exclusive transaction behaviour.
	// An exclusive transaction causes EXCLUSIVE locks to be acquired on all databases.
	// After a BEGIN EXCLUSIVE, no other database connection
	// except for read_uncommitted connections will be able to read the database
	// and no other connection without exception will be able to write the database
	// until the transaction is complete.
	TxLockExclusive = TxLock("BEGIN EXCLUSIVE")
)

// LockingMode defines the database locking mode.
// In NORMAL locking-mode (the default unless overridden at compile-time using SQLITE_DEFAULT_LOCKING_MODE),
// a database connection unlocks the database file at the conclusion of each read or write transaction.
// When the locking-mode is set to EXCLUSIVE,
// the database connection never releases file-locks.
// The first time the database is read in EXCLUSIVE mode,
// a shared lock is obtained and held.
// The first time the database is written, an exclusive lock is obtained and held.
//
// Database locks obtained by a connection in EXCLUSIVE mode may be released
// either by closing the database connection,
// or by setting the locking-mode back to NORMAL.
// Simply setting the locking-mode to NORMAL is not enough -
// locks are not released until the next time the database file is accessed.
//
// There are three reasons to set the locking-mode to EXCLUSIVE.
//
// The application wants to prevent other processes from accessing the database file.
// The number of system calls for filesystem operations is reduced,
// possibly resulting in a small performance increase.
// WAL databases can be accessed in EXCLUSIVE mode without the use of shared memory.
// (Additional information)
// When the locking_mode pragma specifies a particular database, for example:
//
// PRAGMA main.locking_mode=EXCLUSIVE;
// Then the locking mode applies only to the named database.
// If no database name qualifier precedes the "locking_mode" keyword
// then the locking mode is applied to all databases,
// including any new databases added by subsequent ATTACH commands.
//
// The "temp" database (in which TEMP tables and indices are stored)
// and in-memory databases always uses exclusive locking mode.
// The locking mode of temp and in-memory databases cannot be changed.
// All other databases use the normal locking mode by default and are affected by this pragma.
//
// If the locking mode is EXCLUSIVE when first entering WAL journal mode,
// then the locking mode cannot be changed to NORMAL until after exiting WAL journal mode.
// If the locking mode is NORMAL when first entering WAL journal mode,
//then the locking mode can be changed between NORMAL and EXCLUSIVE
// and back again at any time and without needing to exit WAL journal mode.
type LockingMode uint8

const (
	// LockingModeNormal In NORMAL locking-mode
	// (the default unless overridden at compile-time using SQLITE_DEFAULT_LOCKING_MODE),
	// a database connection unlocks the database file at the conclusion
	// of each read or write transaction.
	LockingModeNormal LockingMode = iota

	// LockingModeExclusive When the locking-mode is set to EXCLUSIVE,
	// the database connection never releases file-locks.
	// The first time the database is read in EXCLUSIVE mode,
	// a shared lock is obtained and held.
	// The first time the database is written, an exclusive lock is obtained and held.
	LockingModeExclusive
)

// AutoVacuum defines the auto vacuum status of the database.
// The default setting for auto-vacuum is 0 or "none", unless the SQLITE_DEFAULT_AUTOVACUUM compile-time option is used.
// SQLITE_DEFAULT_AUTOVACUUM can be controlled within the package using
// build tags. See README for more information.
//
// Auto-vacuuming is only possible if the database stores some additional information
// that allows each database page to be traced backwards to its referrer.
// Therefore, auto-vacuuming must be turned on before any tables are created.
// It is not possible to enable or disable auto-vacuum after a table has been created.
type AutoVacuum uint8

const (
	// AutoVacuumNone setting means that auto-vacuum is disabled.
	//
	// When auto-vacuum is disabled and data is deleted data from a database,
	// the database file remains the same size.
	// Unused database file pages are added to a "freelist" and reused for subsequent inserts.
	// So no database file space is lost.
	// However, the database file does not shrink.
	// In this mode the VACUUM command can be used to rebuild the entire database file
	// and thus reclaim unused disk space.
	//
	// The database connection can be changed between full
	// and incremental autovacuum mode at any time.
	// However, changing from "none" to "full" or "incremental"
	// can only occur when the database is new (no tables have yet been created)
	// or by running the VACUUM command. To change auto-vacuum modes,
	// first use the auto_vacuum pragma to set the new desired mode,
	// then invoke the VACUUM command to reorganize the entire database file.
	// To change from "full" or "incremental" back to "none"
	// always requires running VACUUM even on an empty database.
	AutoVacuumNone = AutoVacuum(0)

	// AutoVacuumFull sets auto vacuum of the database to FULL.
	//
	// When the auto-vacuum mode is 1 or "full",
	// the freelist pages are moved to the end of the database file and the database file
	// is truncated to remove the freelist pages at every transaction commit.
	// Note, however, that auto-vacuum only truncates the freelist pages from the file.
	// Auto-vacuum does not defragment the database nor repack individual database pages
	// the way that the VACUUM command does.
	// In fact, because it moves pages around within the file,
	// auto-vacuum can actually make fragmentation worse.
	AutoVacuumFull = AutoVacuum(1)

	// AutoVacuumIncremental sets the auto vacuum of the database to INCREMENTAL.
	//
	// When the value of auto-vacuum is 2 or "incremental"
	// then the additional information needed to do auto-vacuuming is stored
	// in the database file but auto-vacuuming does not occur automatically
	// at each commit as it does with auto_vacuum=full.
	// In incremental mode, the separate incremental_vacuum pragma must be invoked
	//  to cause the auto-vacuum to occur.
	AutoVacuumIncremental = AutoVacuum(2)
)

// JournalMode defines the journal mode associated with the current database connection.
//
// Note that the journal_mode for an in-memory database is either
// MEMORY or OFF and can not be changed to a different value.
// An attempt to change the journal_mode of an in-memory database
// to any setting other than MEMORY or OFF is ignored.
// Note also that the journal_mode cannot be changed while a transaction is active.
type JournalMode string

const (
	// JournalModeDelete  is the normal behavior.
	// In the DELETE mode, the rollback journal is deleted at the conclusion
	// of each transaction.
	// Indeed, the delete operation is the action that causes the transaction to commit.
	// (See the document titled Atomic Commit In SQLite for additional detail.)
	JournalModeDelete JournalMode = "DELETE"

	// JournalModeTruncate commits transactions by truncating the rollback journal
	// to zero-length instead of deleting it.
	// On many systems, truncating a file is much faster
	// than deleting the file since the containing directory does not need to be changed.
	JournalModeTruncate JournalMode = "TRUNCATE"

	// JournalModePersist prevents the rollback journal from being deleted
	// at the end of each transaction.
	// Instead, the header of the journal is overwritten with zeros.
	// This will prevent other database connections from rolling the journal back.
	// The PERSIST journaling mode is useful as an optimization on platforms
	// where deleting or truncating a file is much more expensive
	// than overwriting the first block of a file with zeros.
	// See also: PRAGMA journal_size_limit and SQLITE_DEFAULT_JOURNAL_SIZE_LIMIT.
	JournalModePersist JournalMode = "PERSIST"

	// JournalModeMemory stores the rollback journal in volatile RAM.
	// This saves disk I/O but at the expense of database safety and integrity.
	// If the application using SQLite crashes in the middle of a transaction
	// when the MEMORY journaling mode is set,
	// then the database file will very likely go corrupt.
	JournalModeMemory JournalMode = "MEMORY"

	// JournalModeWAL uses a write-ahead log instead of a rollback journal
	// to implement transactions.
	// The WAL journaling mode is persistent;
	// after being set it stays in effect across multiple database connections
	// and after closing and reopening the database.
	JournalModeWAL JournalMode = "WAL"

	// JournalModeDisabled disables the rollback journal completely.
	// No rollback journal is ever created and hence there is never a rollback journal to delete.
	// The OFF journaling mode disables the atomic commit and rollback capabilities of SQLite. The ROLLBACK command no longer works; it behaves in an undefined way. Applications must avoid using the ROLLBACK command when the journal mode is OFF. If the application crashes in the middle of a transaction when the OFF journaling mode is set, then the database file will very likely go corrupt.
	JournalModeDisabled JournalMode = "OFF"
)

// SecureDelete defines the secure-delete setting.
//
// When secure_delete is on, SQLite overwrites deleted content with zeros.
// The default setting for secure_delete is determined by the SQLITE_SECURE_DELETE
// compile-time option and is normally off.
// The off setting for secure_delete improves performance by reducing
// the number of CPU cycles and the amount of disk I/O.
// Applications that wish to avoid leaving forensic traces after content is deleted
// or updated should enable the secure_delete pragma prior to performing the delete or update,
// or else run VACUUM after the delete or update.
type SecureDelete string

const (
	// SecureDeleteOff disables secure deletion of content.
	SecureDeleteOff = SecureDelete("OFF")

	// SecureDeleteOn will cause SQLite overwrites deleted content with zeros.
	SecureDeleteOn = SecureDelete("ON")

	// SecureDeleteFast defines the "fast" setting for secure_delete (added circa 2017-08-01) is an intermediate setting
	// in between "on" and "off". When secure_delete is set to "fast",
	// SQLite will overwrite deleted content with zeros only if doing so
	// does not increase the amount of I/O. In other words,
	// the "fast" setting uses more CPU cycles but does not use more I/O.
	// This has the effect of purging all old content from b-tree pages,
	// but leaving forensic traces on freelist pages.
	SecureDeleteFast = SecureDelete("FAST")
)

// Synchronous sync setting of the database connection.
type Synchronous uint8

const (
	// SynchronousOff sets synchronous to OFF (0),
	// SQLite continues without syncing as soon as it has handed data off to the operating system.
	// If the application running SQLite crashes, the data will be safe,
	// but the database might become corrupted if the operating system crashes
	// or the computer loses power before that data has been written to the disk surface.
	// On the other hand, commits can be orders of magnitude faster with synchronous OFF.
	SynchronousOff = Synchronous(0)

	// SynchronousNormal sets synchronous to NORMAL (1),
	// the SQLite database engine will still sync at the most critical moments,
	// but less often than in FULL mode.
	// There is a very small (though non-zero) chance that a power failure
	// at just the wrong time could corrupt the database in journal_mode=DELETE
	// on an older filesystem. WAL mode is safe from corruption with synchronous=NORMAL,
	// and probably DELETE mode is safe too on modern filesystems.
	// WAL mode is always consistent with synchronous=NORMAL,
	// but WAL mode does lose durability.
	// A transaction committed in WAL mode with synchronous=NORMAL
	// might roll back following a power loss or system crash.
	// Transactions are durable across application crashes regardless
	// of the synchronous setting or journal mode.
	// The synchronous=NORMAL setting is a good choice for most applications running in WAL mode.
	SynchronousNormal = Synchronous(1)

	// SynchronousFull sets synchronous to FULL (2),
	// the SQLite database engine will use the xSync method of the VFS
	// to ensure that all content is safely written to the disk surface prior to continuing.
	// This ensures that an operating system crash or power failure
	// will not corrupt the database. FULL synchronous is very safe,
	// but it is also slower.
	///FULL is the most commonly used synchronous setting when not in WAL mode.
	SynchronousFull = Synchronous(2)

	// SynchronousExtra is like FULL with the addition that the directory containing
	// a rollback journal is synced after that journal is unlinked to commit a transaction
	// in DELETE mode. EXTRA provides additional durability if the commit
	// is followed closely by a power loss.
	SynchronousExtra = Synchronous(3)
)

// CacheMode defines the shared-cache mode of SQLite.
type CacheMode string

const (
	// CacheModeShared sets the cache mode of SQLite to 'shared'
	CacheModeShared = CacheMode("SHARED")

	// CacheModePrivate sets the cache mode of SQLite to 'private'
	CacheModePrivate = CacheMode("PRIVATE")
)

// Mode defines the open mode of the SQLite database.
type Mode string

const (
	// ModeReadOnly defines SQLITE_OPEN_READONLY for the database connection.
	ModeReadOnly = Mode("RO")

	// ModeReadWrite defines SQLITE_OPEN_READWRITE for the database connection.
	ModeReadWrite = Mode("RW")

	// ModeReadWriteCreate defines SQLITE_OPEN_READWRITE and SQLITE_OPEN_CREATE.
	ModeReadWriteCreate = Mode("RWC")

	// ModeMemory defines mode=memory which will
	// create a pure in-memory database that never reads or writes from disk
	ModeMemory = Mode("MEMORY")
)

// Config is configuration parsed from a DSN string.
// If a new Config is created instead of being parsed from a DSN string,
// the NewConfig function should be used, which sets default values.
// Manual usage is allowed
type Config struct {
	// Mode of the SQLite database
	Mode Mode

	// CacheMode of the SQLite Connection
	Cache CacheMode

	// The immutable parameter is a boolean query parameter that indicates
	// that the database file is stored on read-only media. When immutable is set,
	// SQLite assumes that the database file cannot be changed,
	// even by a process with higher privilege,
	// and so the database is opened read-only and all locking and change detection is disabled.
	// Caution: Setting the immutable property on a database file that
	// does in fact change can result in incorrect query results and/or SQLITE_CORRUPT errors.
	Immutable bool

	// Mutex flag SQLITE_OPEN_MUTEX_NO, SQLITE_OPEN_MUTEX_FULL
	// Defaults to SQLITE_OPEN_MUTEX_FULL
	Mutex int

	// TimeZone location
	TimeZone *time.Location

	// TransactionLock behaviour
	TransactionLock TxLock

	// LockingMode behaviour
	LockingMode LockingMode

	// Authentication holds the UserAuth configuration
	Authentication *Auth

	// AutoVacuum sets the auto vacuum status of the database
	// Defaults to NONE
	AutoVacuum AutoVacuum

	// BusyTimeout defines the time a connection will wait when the
	// connection is BUSY and locked by an other connection.
	// BusyTimeout is defined in milliseconds
	BusyTimeout time.Duration

	// CaseSensitiveLike controls the behaviour of the LIKE operator.
	// Default or disabled the LIKE operation is case-insensitive.
	// When enabling this options behaviour of LIKE will become case-sensitive.
	CaseSensitiveLike bool

	// DeferForeignKeys when enabled will cause the enforcement
	// of all foreign key constraints is delayed until
	// the outermost transaction is committed.
	// The defer_foreign_keys pragma defaults to false
	// so that foreign key constraints are only deferred
	// if they are created as "DEFERRABLE INITIALLY DEFERRED".
	// The defer_foreign_keys pragma is automatically switched off at each COMMIT or ROLLBACK.
	// Hence, the defer_foreign_keys pragma must be separately enabled for each transaction.
	// This pragma is only meaningful if foreign key constraints are enabled, of course.
	DeferForeignKeys bool

	// ForeignKeyConstraints enable or disable the enforcement of foreign key constraints.
	ForeignKeyConstraints bool

	// IgnoreCheckConstraints enables or disables the enforcement of CHECK constraints.
	// The default setting is off, meaning that CHECK constraints are enforced by default.
	IgnoreCheckConstraints bool

	// JournalMode sets the journal mode for databases associated with the current database connection.
	JournalMode JournalMode

	// QueryOnly prevents all changes to the database when set to true.
	QueryOnly bool

	// RecursiveTriggers enable or disable the recursive trigger capability.
	RecursiveTriggers bool

	// SecureDelete enables or disables or sets the secure deletion within the database.
	SecureDelete SecureDelete

	// Synchronous Mode of the database
	Synchronous Synchronous

	// WriteableSchema enables of disables the ability to using UPDATE, INSERT, DELETE
	// Warning: misuse of this pragma can easily result in a corrupt database file.
	WriteableSchema bool
}

// Auth holds the authentication configuration for the SQLite UserAuth module.
type Auth struct {
	// Username for authentication
	Username string

	// Password for authentication
	Password string

	// Salt for encryption
	Salt string

	// CryptEncoder used for the password encryption
	Encoder CryptEncoder
}

// NewConfig creates a new Config and sets default values.
func NewConfig() *Config {
	return &Config{
		Cache:                  CacheModePrivate,
		Immutable:              false,
		Authentication:         &Auth{},
		Mutex:                  SQLITE_OPEN_MUTEX_FULL,
		TransactionLock:        TxLockDeferred,
		LockingMode:            LockingModeNormal,
		AutoVacuum:             AutoVacuumNone,
		CaseSensitiveLike:      false,
		DeferForeignKeys:       false,
		ForeignKeyConstraints:  false,
		IgnoreCheckConstraints: false,
		JournalMode:            JournalModeDelete,
		QueryOnly:              false,
		RecursiveTriggers:      false,
		SecureDelete:           SecureDeleteOff,
		Synchronous:            SynchronousNormal,
		WriteableSchema:        false,
	}
}

// FormatDSN formats the given Config into a DSN string which can be passed to
// the driver.
func (cfg *Config) FormatDSN() string {
	// TODO: FormatDSN
	return ""
}

// Create connection from Configuration
func (cfg *Config) createConnection() (driver.Conn, error) {
	//var db *C.sqlite3


	// name := C.CString(dsn)
	// defer C.free(unsafe.Pointer(name))
	// rv := C._sqlite3_open_v2(name, &db,
	// 	mutex|C.SQLITE_OPEN_READWRITE|C.SQLITE_OPEN_CREATE,
	// 	nil)
	// if rv != 0 {
	// 	return nil, Error{Code: ErrNo(rv)}
	// }
	// if db == nil {
	// 	return nil, errors.New("sqlite succeeded without returning a database")
	// }

	// rv = C.sqlite3_busy_timeout(db, C.int(busyTimeout))
	// if rv != C.SQLITE_OK {
	// 	C.sqlite3_close_v2(db)
	// 	return nil, Error{Code: ErrNo(rv)}
	// }

	// exec := func(s string) error {
	// 	cs := C.CString(s)
	// 	rv := C.sqlite3_exec(db, cs, nil, nil, nil)
	// 	C.free(unsafe.Pointer(cs))
	// 	if rv != C.SQLITE_OK {
	// 		fmt.Printf("-Open-Exec() %d\n", rv)
	// 		return lastError(db)
	// 	}
	// 	return nil
	// }

	// &SQLiteConn{
	// 	db: db,
	// 	tz: cfg.TimeZone,
	// 	txlock: cfg.TransactionLock.String(),
	// }

	return nil, nil
}

// ParseDSN parses the DSN string to a Config
func ParseDSN(dsn string) (cfg *Config, err error) {
	// New default with default values
	cfg = NewConfig()

	pos := strings.IndexRune(dsn, '?')
	if pos >= 1 {
		// Parse Options
		params, err := url.ParseQuery(dsn[pos+1:])
		if err != nil {
			return nil, err
		}

		// Normalize Params
		normalizeParams(params)

		if !strings.HasPrefix(dsn, "file:") {
			dsn = dsn[:pos]
		}

		// Parse Autentication
		if val := params.Get("user"); val != "" {
			cfg.Authentication.Username = val
		}

		if val := params.Get("pass"); val != "" {
			cfg.Authentication.Password = val
		}

		if val := params.Get("salt"); val != "" {
			cfg.Authentication.Salt = val
		}

		if val := params.Get("crypt"); val != "" {
			if cfg.Authentication.Encoder, err = parseCryptEncoder(val, cfg.Authentication.Salt); err != nil {
				return nil, err
			}
		}

		// Parse Multi name options
		// Multi name options are options which has multiple aliases for the same option
		for k := range params {
			// Cache
			if k == "cache" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "shared":
					cfg.Cache = CacheModeShared
				case "private":
					cfg.Cache = CacheModePrivate
				default:
					return nil, fmt.Errorf("Unknown cache mode: %v, expecting value of 'shared, private'", val)
				}
			}

			// Immutable
			if k == "immutable" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.Immutable = false
				case "1", "yes", "true", "on":
					cfg.Immutable = true
				default:
					return nil, fmt.Errorf("Unknown immutable: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Mode
			if k == "mode" {
				val := params.Get(k)
				switch strings.ToUpper(val) {
				case "RO", "RW", "RWC", "MEMORY":
					cfg.Mode = Mode(strings.ToUpper(val))
				default:
					return nil, fmt.Errorf("Unknown mode: %v, expecting value of 'ro, rw, rwc, memory'", val)
				}
			}

			// Timezone
			if k == "tz" || k == "timezone" || k == "loc" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "auto":
					cfg.TimeZone = time.Local
				default:
					cfg.TimeZone, err = time.LoadLocation(val)
					if err != nil {
						return nil, fmt.Errorf("Invalid tz: %v: %v", val, err)
					}
				}
			}

			// Mutex
			if k == "mutex" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "no":
					cfg.Mutex = SQLITE_OPEN_MUTEX_NO
				case "full":
					cfg.Mutex = SQLITE_OPEN_MUTEX_FULL
				default:
					return nil, fmt.Errorf("Invalid mutex: %v, expecting value of 'no, full", val)
				}
			}

			// Transaction Lock
			if k == "txlock" || k == "transaction_lock" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "immediate":
					cfg.TransactionLock = TxLockImmediate
				case "exclusive":
					cfg.TransactionLock = TxLockExclusive
				case "deferred":
					cfg.TransactionLock = TxLockDeferred
				default:
					return nil, fmt.Errorf("Invalid txlock: %v, expecting value of 'deferred, immediate, exclusive'", val)
				}
			}

			// AutoVacuum
			if k == "auto_vacuum" || k == "vacuum" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "none":
					cfg.AutoVacuum = AutoVacuumNone
				case "1", "full":
					cfg.AutoVacuum = AutoVacuumFull
				case "2", "incremental":
					cfg.AutoVacuum = AutoVacuumIncremental
				default:
					return nil, fmt.Errorf("Invalid auto_vacuum: %v, expecting value of '0 NONE 1 FULL 2 INCREMENTAL'", val)
				}
			}

			// Busy Timeout
			if k == "busy_timeout" || k == "timeout" {
				val := params.Get(k)
				iv, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("Invalid busy_timeout: %v: %v", val, err)
				}

				cfg.BusyTimeout, _ = time.ParseDuration(fmt.Sprintf("%dms", iv))
			}

			// Case Sensitive LIKE
			if k == "case_sensitive_like" || k == "cslike" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.CaseSensitiveLike = false
				case "1", "yes", "true", "on":
					cfg.CaseSensitiveLike = true
				default:
					return nil, fmt.Errorf("Invalid case_sensitive_like: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Defer Foreign Keys
			if k == "defer_foreign_keys" || k == "defer_fk" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.DeferForeignKeys = false
				case "1", "yes", "true", "on":
					cfg.DeferForeignKeys = true
				default:
					return nil, fmt.Errorf("Invalid defer_foreign_keys: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Foreign Keys
			if k == "foreign_keys" || k == "fk" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.ForeignKeyConstraints = false
				case "1", "yes", "true", "on":
					cfg.ForeignKeyConstraints = true
				default:
					return nil, fmt.Errorf("Invalid foreign_keys: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Ignore Check Constraints
			if k == "ignore_check_constraints" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.IgnoreCheckConstraints = false
				case "1", "yes", "true", "on":
					cfg.IgnoreCheckConstraints = true
				default:
					return nil, fmt.Errorf("Invalid ignore_check_constraints: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Parse Synchronous before Journal Mode
			// Because if WAL mode is selected for Journal
			// This will change the Synchronous mode
			if k == "synchronous" || k == "sync" {
				val := params.Get(k)
				switch strings.ToUpper(val) {
				case "0", "OFF":
					cfg.Synchronous = SynchronousOff
				case "1", "NORMAL":
					cfg.Synchronous = SynchronousNormal
				case "2", "FULL":
					cfg.Synchronous = SynchronousFull
				case "3", "EXTRA":
					cfg.Synchronous = SynchronousExtra
				default:
					return nil, fmt.Errorf("Invalid synchronous: %v, expecting value of '0 OFF 1 NORMAL 2 FULL 3 EXTRA'", val)
				}
			}

			// Journal Mode
			if k == "journal_mode" || k == "journal" {
				val := params.Get(k)
				switch strings.ToUpper(val) {
				case "DELETE", "TRUNCATE", "PERSIST", "MEMORY", "OFF":
					cfg.JournalMode = JournalMode(strings.ToUpper(val))
				case "WAL":
					cfg.JournalMode = JournalModeWAL

					// For WAL Mode set Synchronous Mode to 'NORMAL'
					// See https://www.sqlite.org/pragma.html#pragma_synchronous
					cfg.Synchronous = SynchronousNormal
				default:
					return nil, fmt.Errorf("Invalid journal: %v, expecting value of 'DELETE TRUNCATE PERSIST MEMORY WAL OFF'", val)
				}
			}

			// Locking Mode
			if k == "locking_mode" || k == "locking" || k == "lock" {
				val := params.Get(k)
				switch strings.ToUpper(val) {
				case "NORMAL":
					cfg.LockingMode = LockingModeNormal
				case "EXCLUSIVE":
					cfg.LockingMode = LockingModeExclusive
				default:
					return nil, fmt.Errorf("Invalid locking_mode: %v, expecting value of 'NORMAL EXCLUSIVE", val)
				}
			}

			// Query Only
			if k == "query_only" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.QueryOnly = false
				case "1", "yes", "true", "on":
					cfg.QueryOnly = true
				default:
					return nil, fmt.Errorf("Invalid query_only: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Recursive Triggers
			if k == "rt" || k == "recursive_triggers" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.RecursiveTriggers = false
				case "1", "yes", "true", "on":
					cfg.RecursiveTriggers = true
				default:
					return nil, fmt.Errorf("Invalid recursive_triggers: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}

			// Secure Delete
			if k == "secure_delete" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.SecureDelete = SecureDeleteOff
				case "1", "yes", "true", "on":
					cfg.SecureDelete = SecureDeleteOn
				case "fast":
					cfg.SecureDelete = SecureDeleteFast
				default:
					return nil, fmt.Errorf("Invalid secure_delete: %v, expecting boolean value of '0 1 false true no yes off on fast'", val)
				}
			}

			if k == "writable_schema" {
				val := params.Get(k)
				switch strings.ToLower(val) {
				case "0", "no", "false", "off":
					cfg.WriteableSchema = false
				case "1", "yes", "true", "on":
					cfg.WriteableSchema = true
				default:
					return nil, fmt.Errorf("Invalid writable_schema: %v, expecting boolean value of '0 1 false true no yes off on'", val)
				}
			}
		}
	}

	return cfg, nil
}

func normalizeParams(params url.Values) {
	for k, v := range params {
		params[strings.ToLower(k)] = v
	}
}

func parseCryptEncoder(crypt, salt string) (CryptEncoder, error) {
	switch strings.ToUpper(crypt) {
	case "SHA1":
		return NewSHA1Encoder(), nil
	case "SSHA1":
		if len(salt) == 0 {
			return nil, fmt.Errorf("crypt=ssha1, requires salt")
		}
		return NewSSHA1Encoder(salt), nil
	case "SHA256":
		return NewSHA256Encoder(), nil
	case "SSHA256":
		if len(salt) == 0 {
			return nil, fmt.Errorf("crypt=ssha256, requires salt")
		}
		return NewSSHA256Encoder(salt), nil
	case "SHA384":
		return NewSHA384Encoder(), nil
	case "SSHA384":
		if len(salt) == 0 {
			return nil, fmt.Errorf("crypt=ssha384, requires salt")
		}
		return NewSSHA384Encoder(salt), nil
	case "SHA512":
		return NewSHA512Encoder(), nil
	case "SSHA512":
		if len(salt) == 0 {
			return nil, fmt.Errorf("crypt=ssha512, requires salt")
		}
		return NewSSHA512Encoder(salt), nil
	default:
		return nil, fmt.Errorf("Unknown crypt encoder provider")
	}
}
