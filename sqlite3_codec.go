// Copyright (C) 2018 CovenantSQL <auxten@covenantsql.io>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_encrypt

package sqlite3

/*
#cgo CFLAGS: -DSQLITE_HAS_CODEC
#cgo CFLAGS: -DSQLITE_TEMP_STORE=2
#cgo CFLAGS: -DSQLCIPHER_CRYPTO_LIBTOMCRYPT
*/
import "C"
