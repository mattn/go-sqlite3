// Copyright (C) 2022 Jonathan Giannuzzi <jonathan@giannuzzi.me>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build sqlcipher || libsqlcipher
// +build sqlcipher libsqlcipher

package sqlite3

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestCipher(t *testing.T) {
	keys := []string{
		// Passphrase with Key Derivation
		"passphrase",
		// Passphrase with Key Derivation starting with a digit
		"1passphrase",
		// Raw Key Data (Without Key Derivation)
		"x'2DD29CA851E7B56E4697B0E1F08507293D761A05CE4D1B628663F411A8086D99'",
		// Raw Key Data with Explicit Salt (Without Key Derivation)
		"x'98483C6EB40B6C31A448C22A66DED3B5E5E8D5119CAC8327B655C8B5C483648101010101010101010101010101010101'",
	}
	for _, key := range keys {
		fname := TempFilename(t)
		uri := "file:" + fname + "?_key=" + key
		db, err := sql.Open("sqlite3", uri)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uri, err)
			continue
		}

		var provider string
		err = db.QueryRow("PRAGMA cipher_provider;").Scan(&provider)
		if err != nil {
			db.Close()
			os.Remove(fname)
			t.Errorf("failed to query cipher_provider for %q: %v", uri, err)
			continue
		}
		if provider == "" {
			db.Close()
			os.Remove(fname)
			t.Errorf("cipher_provider unset for %q - database not encrypted", uri)
			continue
		}

		_, err = db.Exec("CREATE TABLE test (id int)")
		if err != nil {
			db.Close()
			os.Remove(fname)
			t.Errorf("failed creating test table for %q: %v", uri, err)
			continue
		}
		_, err = db.Exec("INSERT INTO test VALUES (1)")
		db.Close()
		if err != nil {
			os.Remove(fname)
			t.Errorf("failed inserting value into test table for %q: %v", uri, err)
			continue
		}

		db, err = sql.Open("sqlite3", "file:"+fname)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", "file:"+fname, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		if err == nil {
			os.Remove(fname)
			t.Errorf("didn't expect to be able to access the encrypted database %q without a passphrase", fname)
			continue
		}

		db, err = sql.Open("sqlite3", "file:"+fname+"?_key=bogus")
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", "file:"+fname+"?_key=bogus", err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		if err == nil {
			os.Remove(fname)
			t.Errorf("didn't expect to be able to access the encrypted database %q with a bogus passphrase", fname)
			continue
		}

		db, err = sql.Open("sqlite3", uri)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uri, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		os.Remove(fname)
		if err != nil {
			t.Errorf("unable to query test table for %q: %v", uri, err)
			continue
		}
	}
}

func TestCipherCompatibility(t *testing.T) {
	db, err := sql.Open("sqlite3", "file::memory:")
	if err != nil {
		t.Fatalf("unable to open in-memory database: %v", err)
	}

	var version string
	err = db.QueryRow("PRAGMA cipher_version;").Scan(&version)
	db.Close()
	if err != nil {
		t.Fatalf("query cipher_version failed: %v", err)
	}

	major, err := strconv.Atoi(version[0:1])
	if err != nil {
		t.Fatalf("parse version major")
	}

	if major < 4 {
		return
	}

	for i := 1; i < major; i++ {
		fname := TempFilename(t)

		uriNoCompat := fmt.Sprintf("file:%s?_key=passphrase", fname)
		uriCompat := fmt.Sprintf("%s&_cipher_compatibility=%d", uriNoCompat, i)
		uriMigrate := fmt.Sprintf("%s&_cipher_migrate", uriNoCompat)

		// Create database in legacy format
		db, err := sql.Open("sqlite3", uriCompat)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uriCompat, err)
			continue
		}
		_, err = db.Exec("CREATE TABLE test (id int)")
		if err != nil {
			db.Close()
			os.Remove(fname)
			t.Errorf("failed creating test table for %q: %v", uriCompat, err)
			continue
		}
		_, err = db.Exec("INSERT INTO test VALUES (1)")
		db.Close()
		if err != nil {
			os.Remove(fname)
			t.Errorf("failed inserting value into test table for %q: %v", uriCompat, err)
			continue
		}

		// Try to open the database without cipher_compability
		db, err = sql.Open("sqlite3", uriNoCompat)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uriNoCompat, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		if err == nil {
			os.Remove(fname)
			t.Errorf("didn't expect to be able to access the encrypted database %q created with version %d without cipher_compatibility", fname, i)
			continue
		}

		// Open the database with cipher_compatibility
		db, err = sql.Open("sqlite3", uriCompat)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uriCompat, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		if err != nil {
			os.Remove(fname)
			t.Errorf("unable to query test table for %q: %v", uriCompat, err)
			continue
		}

		// Migrate the database
		db, err = sql.Open("sqlite3", uriMigrate)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uriMigrate, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		if err != nil {
			os.Remove(fname)
			t.Errorf("unable to query test table for %q: %v", uriMigrate, err)
			continue
		}

		// Try to open the database with cipher_compatibility to validate migration
		db, err = sql.Open("sqlite3", uriCompat)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uriCompat, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		if err == nil {
			os.Remove(fname)
			t.Errorf("didn't expect to be able to access the encrypted database %q after migration with cipher_compatibility", fname)
			continue
		}

		// Open the database without cipher_compatibility to validate migration
		db, err = sql.Open("sqlite3", uriNoCompat)
		if err != nil {
			os.Remove(fname)
			t.Errorf("sql.Open(\"sqlite3\", %q): %v", uriNoCompat, err)
			continue
		}
		_, err = db.Exec("SELECT id FROM test")
		db.Close()
		os.Remove(fname)
		if err != nil {
			t.Errorf("unable to query test table for %q: %v", uriNoCompat, err)
			continue
		}
	}
}
