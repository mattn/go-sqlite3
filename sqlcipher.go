// Copyright (C) 2022 Jonathan Giannuzzi <jonathan@giannuzzi.me>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlcipher

package sqlite3

/*
#cgo CFLAGS: -DUSE_SQLCIPHER
#cgo CFLAGS: -DSQLITE_HAS_CODEC
#cgo CFLAGS: -DSQLITE_TEMP_STORE=2
#cgo !darwin LDFLAGS: -lcrypto
#cgo !darwin CFLAGS: -DSQLCIPHER_CRYPTO_OPENSSL
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security
#cgo darwin CFLAGS: -DSQLCIPHER_CRYPTO_CC
*/
import "C"
