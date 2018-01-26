package sqlite3

import (
	"fmt"
	"testing"
)

func TestLogSafeConfigure_Panic(t *testing.T) {
	driver := &SQLiteDriver{}
	conn, err := driver.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	want := fmt.Sprintf(
		"failed to initialize SQLite logging: %s (%d)",
		errorString(Error{Code: ErrMisuse}), ErrMisuse,
	)

	defer func() {
		got := recover()
		if got != want {
			t.Errorf("expected panic\n%q\ngot\n%q", want, got)
		}
	}()
	logSafeConfigure()
}
