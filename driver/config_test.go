// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDSN(t *testing.T) {
	// URI
	uriCases := map[string]*Config{
		"file:test.db": &Config{
			Database: "file:test.db",
		},
		"file::memory:": &Config{
			Database: "file::memory:",
		},
		"test.db": &Config{
			Database: "test.db",
		},
		":memory:": &Config{
			Database: ":memory:",
		},
		"test.db?%35%2%%43?test=false": nil,
	}

	for dsn, c := range uriCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal(err)
		}
		if c != nil {
			if cfg.Database != c.Database {
				t.Fatalf("Failed to parse database uri; expected: %s, got: %s", c.Database, cfg.Database)
			}
		}
	}

	// Auth
	authCases := map[string]*Config{
		"test.db?user=admin&pass=admin&salt=test": &Config{
			Authentication: &Auth{
				Username: "admin",
				Password: "admin",
				Salt:     "test",
			},
		},
	}

	for dsn, c := range authCases {
		cfg, err := ParseDSN(dsn)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Authentication.Username != c.Authentication.Username {
			t.Fatalf("Failed to parse 'user'; expected: %s, got %s", c.Authentication.Username, cfg.Authentication.Username)
		}
		if cfg.Authentication.Password != c.Authentication.Password {
			t.Fatalf("Failed to parse 'pass'; expected: %s, got %s", c.Authentication.Password, cfg.Authentication.Password)
		}
		if cfg.Authentication.Salt != c.Authentication.Salt {
			t.Fatalf("Failed to parse 'salt'; expected: %s, got %s", c.Authentication.Salt, cfg.Authentication.Salt)
		}
	}

	// Crypt
	RegisterCryptEncoder(NewSSHA1Encoder("salt"))
	RegisterCryptEncoder(NewSSHA256Encoder("salt"))
	RegisterCryptEncoder(NewSSHA384Encoder("salt"))
	RegisterCryptEncoder(NewSSHA512Encoder("salt"))

	cryptCases := map[string]*Config{
		"test.db?crypt=auto": nil,
		"test.db?crypt=sha1": &Config{
			Authentication: &Auth{
				Encoder: NewSHA1Encoder(),
			},
		},
		"test.db?crypt=ssha1": nil,
		"test.db?crypt=ssha1&salt=salt": &Config{
			Authentication: &Auth{
				Salt:    "salt",
				Encoder: NewSSHA1Encoder("salt"),
			},
		},
		"test.db?crypt=sha256": &Config{
			Authentication: &Auth{
				Encoder: NewSHA256Encoder(),
			},
		},
		"test.db?crypt=ssha256": nil,
		"test.db?crypt=ssha256&salt=salt": &Config{
			Authentication: &Auth{
				Salt:    "salt",
				Encoder: NewSSHA256Encoder("salt"),
			},
		},
		"test.db?crypt=sha384": &Config{
			Authentication: &Auth{
				Encoder: NewSHA384Encoder(),
			},
		},
		"test.db?crypt=ssha384": nil,
		"test.db?crypt=ssha384&salt=salt": &Config{
			Authentication: &Auth{
				Salt:    "salt",
				Encoder: NewSSHA384Encoder("salt"),
			},
		},
		"test.db?crypt=sha512": &Config{
			Authentication: &Auth{
				Encoder: NewSHA512Encoder(),
			},
		},
		"test.db?crypt=ssha512": nil,
		"test.db?crypt=ssha512&salt=salt": &Config{
			Authentication: &Auth{
				Salt:    "salt",
				Encoder: NewSSHA512Encoder("salt"),
			},
		},
	}

	for dsn, c := range cryptCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'crypt'")
		}
		if c != nil {
			if reflect.TypeOf(cfg.Authentication.Encoder).String() != reflect.TypeOf(c.Authentication.Encoder).String() {
				t.Fatal("Failed to parse 'crypt'")
			}
			if len(cfg.Authentication.Salt) > 0 {
				if cfg.Authentication.Salt != c.Authentication.Salt {
					t.Fatal("Failed to parse: 'salt'")
				}
			}
		}
	}

	// Cache
	cacheCases := map[string]*Config{
		"test.db?cache=shared": &Config{
			Cache: CacheModeShared,
		},
		"test.db?cache=private": &Config{
			Cache: CacheModePrivate,
		},
		"test.db?cache=bogus": nil,
	}

	for dsn, c := range cacheCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'cache'")
		}
		if c != nil {
			if cfg.Cache != c.Cache {
				t.Fatalf("Failed to parse 'cache'; expected: %d, got: %d", c.Cache, cfg.Cache)
			}
		}
	}

	// Immutable
	immutableCases := map[string]*Config{
		"test.db?immutable=false": &Config{
			Immutable: false,
		},
		"test.db?immutable=true": &Config{
			Immutable: true,
		},
		"test.db?immutable=active": nil,
	}

	for dsn, c := range immutableCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'immutable'")
		}
		if c != nil {
			if cfg.Immutable != c.Immutable {
				t.Fatalf("Failed to parse 'immutable'; expected: %t, got: %t", c.Immutable, cfg.Immutable)
			}
		}
	}

	// Mode
	modeCases := map[string]*Config{
		"test.db?mode=ro": &Config{
			Mode: ModeReadOnly,
		},
		"test.db?mode=rw": &Config{
			Mode: ModeReadWrite,
		},
		"test.db?mode=rwc": &Config{
			Mode: ModeReadWriteCreate,
		},
		"test.db?mode=memory": &Config{
			Mode: ModeMemory,
		},
		"test.db?mode=full": nil,
	}

	for dsn, c := range modeCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'mode'")
		}
		if c != nil {
			if cfg.Mode != c.Mode {
				t.Fatalf("Failed to parse 'mode'; expected: %d, got: %d", c.Mode, cfg.Mode)
			}
		}
	}

	// Mutex
	mutexCases := map[string]*Config{
		"test.db?mutex=no": &Config{
			Mutex: MutexNo,
		},
		"test.db?mutex=full": &Config{
			Mutex: MutexFull,
		},
		"test.db?mutex=bogus": nil,
	}

	for dsn, c := range mutexCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal(err)
		}
		if c != nil {
			if cfg.Mutex != c.Mutex {
				t.Fatalf("Failed to parse 'mutex'; expected: %d, got: %d", c.Mutex, cfg.Mutex)
			}
		}
	}

	// Timezone
	ams, _ := time.LoadLocation("Europe/Amsterdam")
	tzCases := map[string]*Config{
		"test.db?tz=auto": &Config{
			TimeZone: time.Local,
		},
		"test.db?tz=Europe/Amsterdam": &Config{
			TimeZone: ams,
		},
		"test.db?tz=Atlantis": nil,
	}

	for dsn, c := range tzCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal(err)
		}
		if c != nil {
			if cfg.TimeZone.String() != c.TimeZone.String() {
				t.Fatal("Failed to parse timezone")
			}
		}
	}

	// Transaction Lock
	txLockCases := map[string]*Config{
		"test.db?txlock=deferred": &Config{
			TransactionLock: TxLockDeferred,
		},
		"test.db?txlock=immediate": &Config{
			TransactionLock: TxLockImmediate,
		},
		":memory:?txlock=exclusive": &Config{
			TransactionLock: TxLockExclusive,
		},
		"test.db?txlock=bogus": nil,
	}

	for dsn, c := range txLockCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal(err)
		}
		if c != nil {
			if cfg.TransactionLock != c.TransactionLock {
				t.Fatalf("Failed to parse txlock; expected: %s, got: %s", c.TransactionLock, cfg.Database)
			}
		}
	}

	// Auto Vacuum
	vacuumCases := map[string]*Config{
		"test.db?vacuum=none": &Config{
			AutoVacuum: AutoVacuumNone,
		},
		"test.db?vacuum=full": &Config{
			AutoVacuum: AutoVacuumFull,
		},
		"test.db?vacuum=incremental": &Config{
			AutoVacuum: AutoVacuumIncremental,
		},
		"test.db?vacuum=bogus": nil,
	}

	for dsn, c := range vacuumCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'autovacuum, vacuum'")
		}
		if c != nil {
			if cfg.AutoVacuum != c.AutoVacuum {
				t.Fatalf("Failed to parse 'autovacuum'; expected: %s, got: %s", c.AutoVacuum, cfg.AutoVacuum)
			}
		}
	}

	// Busy Timeout
	timeoutCases := map[string]*Config{
		"test.db?timeout=5000": &Config{
			BusyTimeout: 5000 * time.Millisecond,
		},
		"test.db?timeout=never": nil,
	}

	for dsn, c := range timeoutCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'timeout'")
		}
		if c != nil {
			if cfg.BusyTimeout != c.BusyTimeout {
				t.Fatalf("Failed to parse 'timeout'; expected: %d, got: %d", c.BusyTimeout, cfg.BusyTimeout)
			}
		}
	}

	// Case sensitive LIKE
	cslikeCases := map[string]*Config{
		"test.db?cslike=false": &Config{
			CaseSensitiveLike: false,
		},
		"test.db?cslike=true": &Config{
			CaseSensitiveLike: true,
		},
		"test.db?cslike=active": nil,
	}

	for dsn, c := range cslikeCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'cslike'")
		}
		if c != nil {
			if cfg.CaseSensitiveLike != c.CaseSensitiveLike {
				t.Fatalf("Failed to parse 'cslike'; expected: %t, got: %t", c.CaseSensitiveLike, cfg.CaseSensitiveLike)
			}
		}
	}

	// Defer Foreign Keys
	dfkCases := map[string]*Config{
		"test.db?defer_fk=false": &Config{
			DeferForeignKeys: false,
		},
		"test.db?defer_fk=true": &Config{
			DeferForeignKeys: true,
		},
		"test.db?defer_fk=active": nil,
	}

	for dsn, c := range dfkCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'defer_fk'")
		}
		if c != nil {
			if cfg.DeferForeignKeys != c.DeferForeignKeys {
				t.Fatalf("Failed to parse 'defer_fk'; expected: %t, got: %t", c.DeferForeignKeys, cfg.DeferForeignKeys)
			}
		}
	}

	// Foreign Key
	fkCases := map[string]*Config{
		"test.db?fk=false": &Config{
			ForeignKeyConstraints: false,
		},
		"test.db?fk=true": &Config{
			ForeignKeyConstraints: true,
		},
		"test.db?fk=active": nil,
	}

	for dsn, c := range fkCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'fk'")
		}
		if c != nil {
			if cfg.ForeignKeyConstraints != c.ForeignKeyConstraints {
				t.Fatalf("Failed to parse 'fk'; expected: %t, got: %t", c.ForeignKeyConstraints, cfg.ForeignKeyConstraints)
			}
		}
	}

	// Ignore CHECK constraints
	iCases := map[string]*Config{
		"test.db?ignore_check_constraints=false": &Config{
			IgnoreCheckConstraints: false,
		},
		"test.db?ignore_check_constraints=true": &Config{
			IgnoreCheckConstraints: true,
		},
		"test.db?ignore_check_constraints=active": nil,
	}

	for dsn, c := range iCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'ignore_check_constraints'")
		}
		if c != nil {
			if cfg.IgnoreCheckConstraints != c.IgnoreCheckConstraints {
				t.Fatalf("Failed to parse 'ignore_check_constraints'; expected: %t, got: %t", c.IgnoreCheckConstraints, cfg.IgnoreCheckConstraints)
			}
		}
	}

	// Synchronous
	syncCases := map[string]*Config{
		"test.db?sync=off": &Config{
			Synchronous: SynchronousOff,
		},
		"test.db?sync=normal": &Config{
			Synchronous: SynchronousNormal,
		},
		"test.db?sync=full": &Config{
			Synchronous: SynchronousFull,
		},
		"test.db?sync=extra": &Config{
			Synchronous: SynchronousExtra,
		},
		"test.db?sync=bogus": nil,
	}

	for dsn, c := range syncCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'sync'")
		}
		if c != nil {
			if cfg.Synchronous != c.Synchronous {
				t.Fatalf("Failed to parse 'sync'; expected: %s, got: %s", c.Synchronous, cfg.Synchronous)
			}
		}
	}

	// Journal Mode
	journalCases := map[string]*Config{
		"test.db?journal=delete": &Config{
			JournalMode: JournalModeDelete,
		},
		"test.db?journal=truncate": &Config{
			JournalMode: JournalModeTruncate,
		},
		"test.db?journal=persist": &Config{
			JournalMode: JournalModePersist,
		},
		"test.db?journal=memory": &Config{
			JournalMode: JournalModeMemory,
		},
		"test.db?journal=off": &Config{
			JournalMode: JournalModeOff,
		},
		"test.db?journal=wal": &Config{
			JournalMode: JournalModeWAL,
			Synchronous: SynchronousNormal,
		},
		"test.db?journal=auto": nil,
	}

	for dsn, c := range journalCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'journal'")
		}
		if c != nil {
			if cfg.JournalMode != c.JournalMode {
				t.Fatalf("Failed to parse 'journal'; expected: %s, got: %s", c.JournalMode, cfg.JournalMode)
			} else {
				if c.JournalMode == JournalModeWAL {
					if cfg.Synchronous != c.Synchronous {
						t.Fatal("Failed to auto adjust Synchronous mode to normal")
					}
				}
			}
		}
	}

	// Locking Mode
	lockingModeCases := map[string]*Config{
		"test.db?lock=normal": &Config{
			LockingMode: LockingModeNormal,
		},
		"test.db?lock=exclusive": &Config{
			LockingMode: LockingModeExclusive,
		},
		"test.db?lock=auto": nil,
	}

	for dsn, c := range lockingModeCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'locking_mode'")
		}
		if c != nil {
			if cfg.LockingMode != c.LockingMode {
				t.Fatalf("Failed to parse 'locking_mode'; expected: %s, got: %s", c.LockingMode, cfg.LockingMode)
			}
		}
	}

	// Query Only
	qCases := map[string]*Config{
		"test.db?query_only=false": &Config{
			QueryOnly: false,
		},
		"test.db?query_only=true": &Config{
			QueryOnly: true,
		},
		"test.db?query_only=active": nil,
	}

	for dsn, c := range qCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'query_only'")
		}
		if c != nil {
			if cfg.QueryOnly != c.QueryOnly {
				t.Fatalf("Failed to parse 'query_only'; expected: %t, got: %t", c.QueryOnly, cfg.QueryOnly)
			}
		}
	}

	// Recursive Triggers
	rtCases := map[string]*Config{
		"test.db?rt=false": &Config{
			RecursiveTriggers: false,
		},
		"test.db?rt=true": &Config{
			RecursiveTriggers: true,
		},
		"test.db?rt=active": nil,
	}

	for dsn, c := range rtCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'recursive_triggers'")
		}
		if c != nil {
			if cfg.RecursiveTriggers != c.RecursiveTriggers {
				t.Fatalf("Failed to parse 'recursive_triggers'; expected: %t, got: %t", c.RecursiveTriggers, cfg.RecursiveTriggers)
			}
		}
	}

	// Secure Delete
	scCases := map[string]*Config{
		"test.db?secure_delete=off": &Config{
			SecureDelete: SecureDeleteOff,
		},
		"test.db?secure_delete=on": &Config{
			SecureDelete: SecureDeleteOn,
		},
		"test.db?secure_delete=fast": &Config{
			SecureDelete: SecureDeleteFast,
		},
		"test.db?secure_delete=auto": nil,
	}

	for dsn, c := range scCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'secure_delete'")
		}
		if c != nil {
			if cfg.SecureDelete != c.SecureDelete {
				t.Fatalf("Failed to parse 'secure_delete'; expected: %s, got: %s", c.SecureDelete, cfg.SecureDelete)
			}
		}
	}

	// Writable Schema
	wsCases := map[string]*Config{
		"test.db?writable_schema=false": &Config{
			WriteableSchema: false,
		},
		"test.db?writable_schema=true": &Config{
			WriteableSchema: true,
		},
		"test.db?writable_schema=active": nil,
	}

	for dsn, c := range wsCases {
		cfg, err := ParseDSN(dsn)
		if err != nil && c != nil {
			t.Fatal("Failed to parse 'writable_schema'")
		}
		if c != nil {
			if cfg.WriteableSchema != c.WriteableSchema {
				t.Fatalf("Failed to parse 'writable_schema'; expected: %t, got: %t", c.WriteableSchema, cfg.WriteableSchema)
			}
		}
	}
}

func TestFormatDSN(t *testing.T) {
	ams, _ := time.LoadLocation("Europe/Amsterdam")

	cases := map[string]*Config{
		"file::memory:?cache=private&mode=rwc&mutex=full": NewConfig(),
		"file:test.db?mode=ro&mutex=no": &Config{
			Database: "test.db",
			Mode:     ModeReadOnly,
			Mutex:    MutexNo,
		},
		"file:test.db": &Config{
			Database: "test.db",
		},
		"file:test.db?cache=private&cslike=true&defer_fk=true&fk=true&ignore_check_contraints=true&immutable=true&mode=ro&mutex=full&query_only=true&recursive_triggers=true&writable_schema=true": &Config{
			Database:               "test.db",
			Mode:                   ModeReadOnly,
			Mutex:                  MutexFull,
			Cache:                  CacheModePrivate,
			Immutable:              true,
			CaseSensitiveLike:      true,
			DeferForeignKeys:       true,
			ForeignKeyConstraints:  true,
			IgnoreCheckConstraints: true,
			QueryOnly:              true,
			RecursiveTriggers:      true,
			WriteableSchema:        true,
		},
		"file:test.db?cache=shared&mode=rw&mutex=full&tz=auto": &Config{
			Database: "test.db",
			Mode:     ModeReadWrite,
			Mutex:    MutexFull,
			Cache:    CacheModeShared,
			TimeZone: time.Local,
		},
		"file::memory:?cache=private&mode=memory&mutex=full&tz=Europe%2FAmsterdam": &Config{
			Database: ":memory:",
			Mode:     ModeMemory,
			Mutex:    MutexFull,
			Cache:    CacheModePrivate,
			TimeZone: ams,
		},
		"file:test.db?cache=private&mode=ro&mutex=full&txlock=immediate": &Config{
			Database:        "test.db",
			Mode:            ModeReadOnly,
			Mutex:           MutexFull,
			Cache:           CacheModePrivate,
			TransactionLock: TxLockImmediate,
		},
		"file:test.db?cache=private&mode=ro&mutex=full&txlock=exclusive": &Config{
			Database:        "test.db",
			Mode:            ModeReadOnly,
			Mutex:           MutexFull,
			Cache:           CacheModePrivate,
			TransactionLock: TxLockExclusive,
		},
		"file:test.db?cache=private&lock=exclusive&mode=ro&mutex=full": &Config{
			Database:    "test.db",
			Mode:        ModeReadOnly,
			Mutex:       MutexFull,
			Cache:       CacheModePrivate,
			LockingMode: LockingModeExclusive,
		},
		"file:test.db?cache=private&mode=ro&mutex=full&vacuum=full": &Config{
			Database:   "test.db",
			Mode:       ModeReadOnly,
			Mutex:      MutexFull,
			Cache:      CacheModePrivate,
			AutoVacuum: AutoVacuumFull,
		},
		"file:test.db?cache=private&journal=wal&mode=ro&mutex=full": &Config{
			Database:    "test.db",
			Mode:        ModeReadOnly,
			Mutex:       MutexFull,
			Cache:       CacheModePrivate,
			JournalMode: JournalModeWAL,
		},
		"file:test.db?cache=private&mode=ro&mutex=full&secure_delete=fast": &Config{
			Database:     "test.db",
			Mode:         ModeReadOnly,
			Mutex:        MutexFull,
			Cache:        CacheModePrivate,
			SecureDelete: SecureDeleteFast,
		},
		"file:test.db?cache=private&mode=ro&mutex=full&sync=extra": &Config{
			Database:    "test.db",
			Mode:        ModeReadOnly,
			Mutex:       MutexFull,
			Cache:       CacheModePrivate,
			Synchronous: SynchronousExtra,
		},
		"file:test.db?cache=private&crypt=sha1&mode=ro&mutex=full&pass=admin&salt=salt&user=admin": &Config{
			Database: "test.db",
			Mode:     ModeReadOnly,
			Mutex:    MutexFull,
			Cache:    CacheModePrivate,
			Authentication: &Auth{
				Username: "admin",
				Password: "admin",
				Salt:     "salt",
				Encoder:  NewSHA1Encoder(),
			},
		},
	}

	for dsn, cfg := range cases {
		if cfg.FormatDSN() != dsn {
			t.Fatalf("Failed to format DSN; expected: %s, got: %s", dsn, cfg.FormatDSN())
		}
	}
}
