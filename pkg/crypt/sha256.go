// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha256"

// Implementation Enforcer
var (
	_ Encoder = (*sha256Encoder)(nil)
)

type sha256Encoder struct{}

func (e *sha256Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha256.Sum256(pass)
	return h[:]
}

// NewSHA256Encoder returns a new SHA256 Encoder.
func NewSHA256Encoder() Encoder {
	return &sha256Encoder{}
}
