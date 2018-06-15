// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha512"

// Implementation Enforcer
var (
	_ Encoder = (*sha384Encoder)(nil)
)

type sha384Encoder struct{}

func (e *sha384Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha512.Sum384(pass)
	return h[:]
}

// NewSHA384Encoder returns a new SHA384 Encoder.
func NewSHA384Encoder() Encoder {
	return &sha384Encoder{}
}
