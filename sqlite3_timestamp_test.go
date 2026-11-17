// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo

package sqlite3

import (
	"database/sql"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestTimestampWithTimeZone(t *testing.T) {
	const input = "2024-02-03T06:35:06.123456789+02:30"
	for _, zone := range []string{"UTC", "Asia/Tokyo"} {
		for _, decl := range []string{"TIMESTAMP WITH TIME ZONE", `"TiMeStAmP  WiTh TiMe ZoNe"`, "TIMESTAMP", "TEXT", `" DATE "`} {
			t.Run(zone+"/"+decl, func(t *testing.T) {
				db, err := sql.Open("sqlite3", ":memory:?_loc="+url.QueryEscape(zone))
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				db.SetMaxOpenConns(1)
				if _, err := db.Exec("CREATE TABLE times (v " + decl + ")"); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec("INSERT INTO times VALUES (?)", input); err != nil {
					t.Fatal(err)
				}
				rows, err := db.Query("SELECT v FROM times")
				if err != nil {
					t.Fatal(err)
				}
				columns, err := rows.ColumnTypes()
				if err != nil {
					rows.Close()
					t.Fatal(err)
				}
				wantTime := decl != "TEXT" && decl != `" DATE "`
				wantScanType := reflect.TypeOf(sql.NullTime{})
				if decl == "TEXT" {
					wantScanType = reflect.TypeOf(sql.NullString{})
				} else if decl == `" DATE "` {
					wantScanType = reflect.TypeOf(new(any))
				}
				if columns[0].ScanType() != wantScanType {
					rows.Close()
					t.Fatalf("ScanType for %q: %v", decl, columns[0].ScanType())
				}
				if err := rows.Close(); err != nil {
					t.Fatal(err)
				}
				if !wantTime {
					var got string
					if err := db.QueryRow("SELECT v FROM times").Scan(&got); err != nil || got != input {
						t.Fatalf("TEXT got %q, err=%v", got, err)
					}
					return
				}
				loc, err := time.LoadLocation(zone)
				if err != nil {
					t.Fatal(err)
				}
				want := time.Date(2024, time.February, 3, 4, 5, 6, 123456789, time.UTC).In(loc)
				var got time.Time
				if err := db.QueryRow("SELECT v FROM times").Scan(&got); err != nil || !got.Equal(want) || got.Location().String() != zone {
					t.Fatalf("got %v, err=%v, want %v in %s", got, err, want, zone)
				}
			})
		}
	}
}
