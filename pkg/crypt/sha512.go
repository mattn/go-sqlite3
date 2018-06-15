// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha512"

// Implementation Enforcer
var (
	_ Encoder = (*sha512Encoder)(nil)
)

type sha512Encoder struct{}

func (e *sha512Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha512.Sum512(pass)
	return h[:]
}

// NewSHA512Encoder returns a new SHA512 Encoder.
func NewSHA512Encoder() Encoder {
	return &sha512Encoder{}
}
