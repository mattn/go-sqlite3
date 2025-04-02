// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build solaris
// +build solaris

package sqlite3

/*
#cgo CFLAGS: -D__EXTENSIONS__=1
#cgo LDFLAGS: -lc -lcrypto -lsqlcipher

#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h" // Use amalgamation if enabled
#else
#include <sqlcipher/sqlite3.h> // Use system-provided SQLCipher
#endif
*/
import "C"
