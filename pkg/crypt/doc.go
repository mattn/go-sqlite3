// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// package crypt provides several different implementations for the
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
package crypt
