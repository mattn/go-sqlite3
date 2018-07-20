// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

const (
	PRAGMA_SSE_KEY                 = "key"
	PRAGMA_AUTO_VACUUM             = "auto_vacuum"
	PRAGMA_CASE_SENSITIVE_LIKE     = "case_sensitive_like"
	PRAGMA_DEFER_FOREIGN_KEYS      = "defer_foreign_keys"
	PRAGMA_FOREIGN_KEYS            = "foreign_keys"
	PRAGMA_IGNORE_CHECK_CONTRAINTS = "ignore_check_constraints"
	PRAGMA_JOURNAL_MODE            = "journal_mode"
	PRAGMA_LOCKING_MODE            = "locking_mode"
	PRAGMA_QUERY_ONLY              = "query_only"
	PRAGMA_RECURSIVE_TRIGGERS      = "recursive_triggers"
	PRAGMA_SECURE_DELETE           = "secure_delete"
	PRAGMA_SYNCHRONOUS             = "synchronous"
	PRAGMA_WRITABLE_SCHEMA         = "writable_schema"
)
