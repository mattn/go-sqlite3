// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.

// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_introspect
// +build sqlite_introspect

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_INTROSPECTION_PRAGMAS
#cgo LDFLAGS: -lcrypto -lsqlcipher
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h" // Use amalgamation if enabled
#else
#include <sqlcipher/sqlite3.h> // Use system-provided SQLCipher
#endif
*/
import "C"
