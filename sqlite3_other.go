// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build !windows
// +build !windows

package sqlite3

/*
#cgo CFLAGS: -I.
#cgo linux LDFLAGS: -ldl -lcrypto -lsqlcipher
#cgo linux,ppc LDFLAGS: -lpthread -lcrypto -lsqlcipher
#cgo linux,ppc64 LDFLAGS: -lpthread -lcrypto -lsqlcipher
#cgo linux,ppc64le LDFLAGS: -lpthread -lcrypto -lsqlcipher

#ifndef USE_LIBSQLITE3
#include "sqlite3-binding.h" // Use amalgamation if enabled
#else
#include <sqlcipher/sqlite3.h> // Use system-provided SQLCipher
#endif
*/
import "C"
