// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha1"

// Implementation Enforcer
var (
	_ Encoder = (*ssha1Encoder)(nil)
)

type ssha1Encoder struct {
	salt string
}

func (e *ssha1Encoder) Encode(pass []byte, hash interface{}) []byte {
	s := []byte(e.salt)
	p := append(pass, s...)
	h := sha1.Sum(p)
	return h[:]
}

func (e *ssha1Encoder) Salt() string {
	return e.salt
}

// NewSSHA1Encoder returns a new salted SHA1 Encoder.
func NewSSHA1Encoder(salt string) Encoder {
	return &ssha1Encoder{
		salt: salt,
	}
}
