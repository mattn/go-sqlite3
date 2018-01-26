// +build go1.7

package sqlite3

import (
	"testing"

	"github.com/mpvl/subtest"
)

func TestPragmaSetAndCheck_InvalidValueType(t *testing.T) {
	const want = "unsupported value type for pragma statement"
	defer func() {
		got := recover()
		if got != want {
			t.Errorf("expected\n%q\ngot\n%q", want, got)
		}
	}()
	pragmaSetAndCheck(nil, "foo", []byte("bar"))
}

func TestPragmaSetAndCheck_Failure(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value interface{}
		want  string
	}{
		{
			name:  "query returns no rows",
			key:   "foo",
			value: "bar",
			want:  "can't fetch rows for 'PRAGMA foo=bar': EOF",
		},
		{
			name:  "query returns non-int64 row",
			key:   "journal_mode",
			value: int64(1),
			want:  "query 'PRAGMA journal_mode=1' returned a non-int64 row",
		},
		{
			name:  "query returns non-byte row",
			key:   "journal_size_limit",
			value: "foo",
			want:  "query 'PRAGMA journal_size_limit=foo' returned a non-byte row",
		},
	}
	for _, c := range cases {
		subtest.Run(t, c.name, func(t *testing.T) {
			driver := &SQLiteDriver{}
			conn, err := driver.Open(":memory:")
			if err != nil {
				t.Fatalf("can't open memory connection: %v", err)
			}
			defer conn.Close()

			err = pragmaSetAndCheck(conn.(*SQLiteConn), c.key, c.value)
			if err == nil {
				t.Fatal("no error returned")
			}
			got := err.Error()
			if got != c.want {
				t.Errorf("got error '%s', want '%s'", got, c.want)
			}
		})
	}
}
