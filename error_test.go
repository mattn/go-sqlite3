package sqlite3

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestFailures(t *testing.T) {
	dirName, err := ioutil.TempDir("", "sqlite3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirName)

	dbFileName := path.Join(dirName, "test.db")
	f, err := os.Create(dbFileName)
	if err != nil {
		t.Error(err)
	}
	f.Write([]byte{1, 2, 3, 4, 5})
	f.Close()

	db, err := sql.Open("sqlite3", dbFileName)
	if err == nil {
		_, err = db.Exec("drop table foo")
	}
	if err != ErrNotADB {
		t.Error("wrong error code for corrupted DB")
	}
	if err.Error() == "" {
		t.Error("wrong error string for corrupted DB")
	}
	db.Close()
}
