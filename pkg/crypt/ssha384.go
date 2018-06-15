// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package crypt

import "crypto/sha512"

// Implementation Enforcer
var (
	_ Encoder = (*ssha384Encoder)(nil)
)

type ssha384Encoder struct {
	salt string
}

func (e *ssha384Encoder) Encode(pass []byte, hash interface{}) []byte {
	s := []byte(e.salt)
	p := append(pass, s...)
	h := sha512.Sum384(p)
	return h[:]
}

func (e *ssha384Encoder) Salt() string {
	return e.salt
}

// NewSSHA384Encoder returns a new salted SHA384 Encoder.
func NewSSHA384Encoder(salt string) Encoder {
	return &ssha384Encoder{
		salt: salt,
	}
}
