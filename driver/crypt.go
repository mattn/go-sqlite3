// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sqlite3

// The crypt functions provides several different implementations for the
// default embedded sqlite_crypt function.
// This function is uses a ceasar-cypher by default
// and is used within the UserAuthentication module to encode
// the password.
//
// The provided functions can be used as an overload to the sqlite_crypt
// function through the use of the RegisterFunc on the connection.
//
// Because the functions can serv a purpose to an end-user
// without using the UserAuthentication module
// the functions are default compiled in.
//
// From SQLITE3 - user-auth.txt
// The sqlite_user.pw field is encoded by a built-in SQL function
// "sqlite_crypt(X,Y)".  The two arguments are both BLOBs.  The first argument
// is the plaintext password supplied to the sqlite3_user_authenticate()
// interface.  The second argument is the sqlite_user.pw value and is supplied
// so that the function can extract the "salt" used by the password encoder.
// The result of sqlite_crypt(X,Y) is another blob which is the value that
// ends up being stored in sqlite_user.pw.  To verify credentials X supplied
// by the sqlite3_user_authenticate() routine, SQLite runs:
//
//     sqlite_user.pw == sqlite_crypt(X, sqlite_user.pw)
//
// To compute an appropriate sqlite_user.pw value from a new or modified
// password X, sqlite_crypt(X,NULL) is run.  A new random salt is selected
// when the second argument is NULL.
//
// The built-in version of of sqlite_crypt() uses a simple Ceasar-cypher
// which prevents passwords from being revealed by searching the raw database
// for ASCII text, but is otherwise trivally broken.  For better password
// security, the database should be encrypted using the SQLite Encryption
// Extension or similar technology.  Or, the application can use the
// sqlite3_create_function() interface to provide an alternative
// implementation of sqlite_crypt() that computes a stronger password hash,
// perhaps using a cryptographic hash function like SHA1.

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

// Force Implementation
var (
	_ CryptEncoder       = (*sha1Encoder)(nil)
	_ CryptEncoder       = (*sha256Encoder)(nil)
	_ CryptEncoder       = (*sha384Encoder)(nil)
	_ CryptEncoder       = (*sha512Encoder)(nil)
	_ CryptSaltedEncoder = (*ssha1Encoder)(nil)
	_ CryptSaltedEncoder = (*ssha256Encoder)(nil)
	_ CryptSaltedEncoder = (*ssha384Encoder)(nil)
	_ CryptSaltedEncoder = (*ssha512Encoder)(nil)
)

var (
	cryptEncoders map[string]CryptEncoder
)

// RegisterCryptEncoder will register a CryptEncoder to the sqlite3 driver
// which will automatically be found when used from a DSN string
func RegisterCryptEncoder(e CryptEncoder) {
	cryptEncoders[e.String()] = e
}

func init() {
	cryptEncoders = make(map[string]CryptEncoder, 0)

	// Register Default Encoders
	RegisterCryptEncoder(NewSHA1Encoder())
	RegisterCryptEncoder(NewSHA256Encoder())
	RegisterCryptEncoder(NewSHA384Encoder())
	RegisterCryptEncoder(NewSHA512Encoder())
}

// #############################################################################

// CryptEncoder provides the interface for implementing
// a sqlite_crypt encoder.
type CryptEncoder interface {
	fmt.Stringer
	Encode(pass []byte, hash interface{}) []byte
}

// CryptSaltedEncoder provides the interface for a encoder
// to return its configured salt.
type CryptSaltedEncoder interface {
	CryptEncoder
	Salt() string
}

// #############################################################################

type sha1Encoder struct{}

func (e *sha1Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha1.Sum(pass)
	return h[:]
}

func (e *sha1Encoder) String() string {
	return "sha1"
}

// NewSHA1Encoder returns a new SHA1 Encoder.
func NewSHA1Encoder() CryptEncoder {
	return &sha1Encoder{}
}

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

func (e *ssha1Encoder) String() string {
	return "ssha1"
}

// NewSSHA1Encoder returns a new salted SHA1 Encoder.
func NewSSHA1Encoder(salt string) CryptSaltedEncoder {
	return &ssha1Encoder{
		salt: salt,
	}
}

type sha256Encoder struct{}

func (e *sha256Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha256.Sum256(pass)
	return h[:]
}

func (e *sha256Encoder) String() string {
	return "sha256"
}

// NewSHA256Encoder returns a new SHA256 Encoder.
func NewSHA256Encoder() CryptEncoder {
	return &sha256Encoder{}
}

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

func (e *ssha256Encoder) String() string {
	return "ssha256"
}

// NewSSHA256Encoder returns a new salted SHA256 Encoder.
func NewSSHA256Encoder(salt string) CryptSaltedEncoder {
	return &ssha256Encoder{
		salt: salt,
	}
}

type sha384Encoder struct{}

func (e *sha384Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha512.Sum384(pass)
	return h[:]
}

func (e *sha384Encoder) String() string {
	return "sha384"
}

// NewSHA384Encoder returns a new SHA384 Encoder.
func NewSHA384Encoder() CryptEncoder {
	return &sha384Encoder{}
}

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

func (e *ssha384Encoder) String() string {
	return "ssha384"
}

// NewSSHA384Encoder returns a new salted SHA384 Encoder.
func NewSSHA384Encoder(salt string) CryptSaltedEncoder {
	return &ssha384Encoder{
		salt: salt,
	}
}

type sha512Encoder struct{}

func (e *sha512Encoder) Encode(pass []byte, hash interface{}) []byte {
	h := sha512.Sum512(pass)
	return h[:]
}

func (e *sha512Encoder) String() string {
	return "sha512"
}

// NewSHA512Encoder returns a new SHA512 Encoder.
func NewSHA512Encoder() CryptEncoder {
	return &sha512Encoder{}
}

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

func (e *ssha512Encoder) String() string {
	return "ssha512"
}

// NewSSHA512Encoder returns a new salted SHA512 Encoder.
func NewSSHA512Encoder(salt string) CryptSaltedEncoder {
	return &ssha512Encoder{
		salt: salt,
	}
}

// EOF
