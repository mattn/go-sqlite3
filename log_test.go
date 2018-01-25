package sqlite3_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/CanonicalLtd/go-sqlite3"
)

func TestLogConfig_SQLiteUsesConfiguredLogFunc(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	f := func(code int, message string) {
		buffer.WriteString(fmt.Sprintf("[%d] %s", code, message))
	}
	sqlite3.LogConfig(f)
	defer sqlite3.LogConfig(nil)

	conn, _ := (&sqlite3.SQLiteDriver{}).Open(":memory:")
	conn.Prepare("crap")

	want := "[1] near \"crap\": syntax error"
	got := buffer.String()
	if got != want {
		t.Errorf("unexpected output:\n want: %v\n  got: %v", want, got)
	}
}

func TestLogConfig_ReturnPreviousLogger(t *testing.T) {
	f1 := func(int, string) {}
	f2 := func(int, string) {}

	if sqlite3.LogConfig(f1) != nil {
		t.Fatal("expected to have no previous logger")
	}
	defer sqlite3.LogConfig(nil)

	if sqlite3.LogConfig(f2) == nil {
		t.Fatal("expected to return previous logger")
	}
}
