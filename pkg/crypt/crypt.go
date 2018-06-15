// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

// Encoder provides the interface for implementing
// a sqlite_crypt encoder.
type Encoder interface {
	Encode(pass []byte, hash interface{}) []byte
}

// Salter provides the interface for a encoder
// to return its configured salt.
type Salter interface {
	Salt() string
}

// EOF
