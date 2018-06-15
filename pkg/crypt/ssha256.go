// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha256"

// Implementation Enforcer
var (
	_ Encoder = (*ssha256Encoder)(nil)
)

type ssha256Encoder struct {
	salt string
}

func (e *ssha256Encoder) Encode(pass []byte, hash interface{}) []byte {
	s := []byte(e.salt)
	p := append(pass, s...)
	h := sha256.Sum256(p)
	return h[:]
}

func (e *ssha256Encoder) Salt() string {
	return e.salt
}

// NewSSHA256Encoder returns a new salted SHA256 Encoder.
func NewSSHA256Encoder(salt string) Encoder {
	return &ssha256Encoder{
		salt: salt,
	}
}
