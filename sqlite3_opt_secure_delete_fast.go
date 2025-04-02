// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlite_secure_delete_fast
// +build sqlite_secure_delete_fast

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_SECURE_DELETE=FAST
#cgo LDFLAGS: -lcrypto -lsqlcipher
#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h" // Use amalgamation if enabled
#else
#include <sqlcipher/sqlite3.h> // Use system-provided SQLCipher
#endif
*/
import "C"
