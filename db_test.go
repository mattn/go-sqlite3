// Copyright 2017 Canonical Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqlite3_test

import (
	"testing"
	"time"

	"github.com/CanonicalLtd/go-sqlite3"
)

func TestDatabaseSize(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	mustExec(conn, "CREATE TABLE test (n INT)", nil)

	if !(sqlite3.DatabaseSize(conn) > 0) {
		t.Errorf("did not return -1, meaning missing file")
	}
}

func TestDatabaseSize_FileDoesNotExists(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	cleanup() // Remove the underlying database file

	if sqlite3.DatabaseSize(conn) != -1 {
		t.Errorf("did not return -1, meaning missing file")
	}
}

func TestDatabaseModTime(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	defer cleanup()

	timestamp1 := sqlite3.DatabaseModTime(conn)

	time.Sleep(10 * time.Millisecond)
	mustExec(conn, "CREATE TABLE test (n INT)", nil)

	timestamp2 := sqlite3.DatabaseModTime(conn)

	if timestamp1.Equal(timestamp2) {
		t.Errorf("modification time did not change")
	}
}

func TestDatabaseModTime_StatError(t *testing.T) {
	conn, cleanup := newFileSQLiteConn()
	cleanup() // Remove the underlying database file

	want := time.Time{}
	got := sqlite3.DatabaseModTime(conn)
	if !got.Equal(want) {
		t.Errorf("expected timestamp\n%q\ngot\n%q", want, got)
	}
}
