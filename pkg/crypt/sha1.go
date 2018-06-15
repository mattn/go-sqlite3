// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha1"

// Implementation Enforcer
var (
	_ Encoder = (*sha1Encoder)(nil)
)

type sha1Encoder struct{}

func (e *sha1Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha1.Sum(pass)
	return h[:]
}

// NewSHA1Encoder returns a new SHA1 Encoder.
func NewSHA1Encoder() Encoder {
	return &sha1Encoder{}
}
