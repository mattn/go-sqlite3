// Copyright (C) 2018 G.J.R. Timmer <gjr.timmer@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build sqlite_userauth

package sqlite3

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestCreateAuthDatabase(t *testing.T) {
	tempFilename := TempFilename(t)
	fmt.Println(tempFilename) // debug
	//defer os.Remove(tempFilename) // Disable for debug

	db, err := sql.Open("sqlite3", "file:"+tempFilename+"?_auth&_auth_user=admin&_auth_pass=admin")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	var i int64
	err = db.QueryRow("SELECT count(type) FROM sqlite_master WHERE type='table' AND name='sqlite_user';").Scan(&i)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sqlite_user exists: %d", i)

	_, err = db.Exec("SELECT auth_user_add('test', 'test', false);", nil)
	if err != nil {
		t.Fatal(err)
	}

}
