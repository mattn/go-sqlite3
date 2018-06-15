// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha512"

// Implementation Enforcer
var (
	_ Encoder = (*ssha512Encoder)(nil)
)

type ssha512Encoder struct {
	salt string
}

func (e *ssha512Encoder) Encode(pass []byte, hash interface{}) []byte {
	s := []byte(e.salt)
	p := append(pass, s...)
	h := sha512.Sum512(p)
	return h[:]
}

func (e *ssha512Encoder) Salt() string {
	return e.salt
}

// NewSSHA384Encoder returns a new salted SHA512 Encoder.
func NewSSHA512Encoder(salt string) Encoder {
	return &ssha512Encoder{
		salt: salt,
	}
}
