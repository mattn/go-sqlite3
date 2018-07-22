// Copyright (C) 2018 The Go-SQLite3 Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build cgo

package sqlite3

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func timezone(t time.Time) string { return t.Format("-07:00") }

func TestTimestamp(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE foo")
	_, err = db.Exec("CREATE TABLE foo(id INTEGER, ts timeSTAMP, dt DATETIME)")
	if err != nil {
		t.Fatal("failed to create table:", err)
	}

	timestamp1 := time.Date(2012, time.April, 6, 22, 50, 0, 0, time.UTC)
	timestamp2 := time.Date(2006, time.January, 2, 15, 4, 5, 123456789, time.UTC)
	timestamp3 := time.Date(2012, time.November, 4, 0, 0, 0, 0, time.UTC)
	tzTest := time.FixedZone("TEST", -9*3600-13*60)
	tests := []struct {
		value    interface{}
		expected time.Time
	}{
		{"nonsense", time.Time{}},
		{"0000-00-00 00:00:00", time.Time{}},
		{time.Time{}.Unix(), time.Time{}},
		{timestamp1, timestamp1},
		{timestamp2.Unix(), timestamp2.Truncate(time.Second)},
		{timestamp2.UnixNano() / int64(time.Millisecond), timestamp2.Truncate(time.Millisecond)},
		{timestamp1.In(tzTest), timestamp1.In(tzTest)},
		{timestamp1.Format("2006-01-02 15:04:05.000"), timestamp1},
		{timestamp1.Format("2006-01-02T15:04:05.000"), timestamp1},
		{timestamp1.Format("2006-01-02 15:04:05"), timestamp1},
		{timestamp1.Format("2006-01-02T15:04:05"), timestamp1},
		{timestamp2, timestamp2},
		{"2006-01-02 15:04:05.123456789", timestamp2},
		{"2006-01-02T15:04:05.123456789", timestamp2},
		{"2006-01-02T05:51:05.123456789-09:13", timestamp2.In(tzTest)},
		{"2012-11-04", timestamp3},
		{"2012-11-04 00:00", timestamp3},
		{"2012-11-04 00:00:00", timestamp3},
		{"2012-11-04 00:00:00.000", timestamp3},
		{"2012-11-04T00:00", timestamp3},
		{"2012-11-04T00:00:00", timestamp3},
		{"2012-11-04T00:00:00.000", timestamp3},
		{"2006-01-02T15:04:05.123456789Z", timestamp2},
		{"2012-11-04Z", timestamp3},
		{"2012-11-04 00:00Z", timestamp3},
		{"2012-11-04 00:00:00Z", timestamp3},
		{"2012-11-04 00:00:00.000Z", timestamp3},
		{"2012-11-04T00:00Z", timestamp3},
		{"2012-11-04T00:00:00Z", timestamp3},
		{"2012-11-04T00:00:00.000Z", timestamp3},
	}
	for i := range tests {
		_, err = db.Exec("INSERT INTO foo(id, ts, dt) VALUES(?, ?, ?)", i, tests[i].value, tests[i].value)
		if err != nil {
			t.Fatal("failed to insert timestamp:", err)
		}
	}

	rows, err := db.Query("SELECT id, ts, dt FROM foo ORDER BY id ASC")
	if err != nil {
		t.Fatal("unable to query foo table:", err)
	}
	defer rows.Close()

	seen := 0
	for rows.Next() {
		var id int
		var ts, dt time.Time

		if err := rows.Scan(&id, &ts, &dt); err != nil {
			t.Error("unable to scan results:", err)
			continue
		}
		if id < 0 || id >= len(tests) {
			t.Error("bad row id: ", id)
			continue
		}
		seen++
		if !tests[id].expected.Equal(ts) {
			t.Errorf("timestamp value for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected, dt)
		}
		if !tests[id].expected.Equal(dt) {
			t.Errorf("datetime value for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected, dt)
		}
		if timezone(tests[id].expected) != timezone(ts) {
			t.Errorf("timezone for id %v (%v) should be %v, not %v", id, tests[id].value,
				timezone(tests[id].expected), timezone(ts))
		}
		if timezone(tests[id].expected) != timezone(dt) {
			t.Errorf("timezone for id %v (%v) should be %v, not %v", id, tests[id].value,
				timezone(tests[id].expected), timezone(dt))
		}
	}

	if seen != len(tests) {
		t.Errorf("expected to see %d rows", len(tests))
	}
}

func TestTimezoneConversion(t *testing.T) {
	zones := []string{"UTC", "US/Central", "US/Pacific", "Local"}
	for _, tz := range zones {
		tempFilename := TempFilename(t)
		defer os.Remove(tempFilename)
		db, err := sql.Open("sqlite3", tempFilename+"?tz="+url.QueryEscape(tz))
		if err != nil {
			t.Fatal("failed to open database:", err)
		}
		defer db.Close()

		_, err = db.Exec("DROP TABLE foo")
		_, err = db.Exec("CREATE TABLE foo(id INTEGER, ts TIMESTAMP, dt DATETIME)")
		if err != nil {
			t.Fatal("failed to create table:", err)
		}

		loc, err := time.LoadLocation(tz)
		if err != nil {
			t.Fatal("failed to load location:", err)
		}

		timestamp1 := time.Date(2012, time.April, 6, 22, 50, 0, 0, time.UTC)
		timestamp2 := time.Date(2006, time.January, 2, 15, 4, 5, 123456789, time.UTC)
		timestamp3 := time.Date(2012, time.November, 4, 0, 0, 0, 0, time.UTC)
		tests := []struct {
			value    interface{}
			expected time.Time
		}{
			{"nonsense", time.Time{}.In(loc)},
			{"0000-00-00 00:00:00", time.Time{}.In(loc)},
			{timestamp1, timestamp1.In(loc)},
			{timestamp1.Unix(), timestamp1.In(loc)},
			{timestamp1.In(time.FixedZone("TEST", -7*3600)), timestamp1.In(loc)},
			{timestamp1.Format("2006-01-02 15:04:05.000"), timestamp1.In(loc)},
			{timestamp1.Format("2006-01-02T15:04:05.000"), timestamp1.In(loc)},
			{timestamp1.Format("2006-01-02 15:04:05"), timestamp1.In(loc)},
			{timestamp1.Format("2006-01-02T15:04:05"), timestamp1.In(loc)},
			{timestamp2, timestamp2.In(loc)},
			{"2006-01-02 15:04:05.123456789", timestamp2.In(loc)},
			{"2006-01-02T15:04:05.123456789", timestamp2.In(loc)},
			{"2012-11-04", timestamp3.In(loc)},
			{"2012-11-04 00:00", timestamp3.In(loc)},
			{"2012-11-04 00:00:00", timestamp3.In(loc)},
			{"2012-11-04 00:00:00.000", timestamp3.In(loc)},
			{"2012-11-04T00:00", timestamp3.In(loc)},
			{"2012-11-04T00:00:00", timestamp3.In(loc)},
			{"2012-11-04T00:00:00.000", timestamp3.In(loc)},
		}
		for i := range tests {
			_, err = db.Exec("INSERT INTO foo(id, ts, dt) VALUES(?, ?, ?)", i, tests[i].value, tests[i].value)
			if err != nil {
				t.Fatal("failed to insert timestamp:", err)
			}
		}

		rows, err := db.Query("SELECT id, ts, dt FROM foo ORDER BY id ASC")
		if err != nil {
			t.Fatal("unable to query foo table:", err)
		}
		defer rows.Close()

		seen := 0
		for rows.Next() {
			var id int
			var ts, dt time.Time

			if err := rows.Scan(&id, &ts, &dt); err != nil {
				t.Error("unable to scan results:", err)
				continue
			}
			if id < 0 || id >= len(tests) {
				t.Error("bad row id: ", id)
				continue
			}
			seen++
			if !tests[id].expected.Equal(ts) {
				t.Errorf("timestamp value for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected, ts)
			}
			if !tests[id].expected.Equal(dt) {
				t.Errorf("datetime value for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected, dt)
			}
			if tests[id].expected.Location().String() != ts.Location().String() {
				t.Errorf("location for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected.Location().String(), ts.Location().String())
			}
			if tests[id].expected.Location().String() != dt.Location().String() {
				t.Errorf("location for id %v (%v) should be %v, not %v", id, tests[id].value, tests[id].expected.Location().String(), dt.Location().String())
			}
		}

		if seen != len(tests) {
			t.Errorf("expected to see %d rows", len(tests))
		}
	}
}

func TestDateTimeLocal(t *testing.T) {
	zone := "Asia/Tokyo"
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename+"?tz="+zone)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	db.Exec("CREATE TABLE foo (dt datetime);")
	db.Exec("INSERT INTO foo VALUES('2015-03-05 15:16:17');")

	row := db.QueryRow("select * from foo")
	var d time.Time
	err = row.Scan(&d)
	if err != nil {
		t.Fatal("failed to scan datetime:", err)
	}
	if d.Hour() == 15 || !strings.Contains(d.String(), "JST") {
		t.Fatal("result should have timezone", d)
	}
	db.Close()

	db, err = sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}

	row = db.QueryRow("select * from foo")
	err = row.Scan(&d)
	if err != nil {
		t.Fatal("failed to scan datetime:", err)
	}
	if d.UTC().Hour() != 15 || !strings.Contains(d.String(), "UTC") {
		t.Fatalf("result should not have timezone %v %v", zone, d.String())
	}

	_, err = db.Exec("DELETE FROM foo")
	if err != nil {
		t.Fatal("failed to delete table:", err)
	}
	dt, err := time.Parse("2006/1/2 15/4/5 -0700 MST", "2015/3/5 15/16/17 +0900 JST")
	if err != nil {
		t.Fatal("failed to parse datetime:", err)
	}
	db.Exec("INSERT INTO foo VALUES(?);", dt)

	db.Close()
	db, err = sql.Open("sqlite3", tempFilename+"?tz="+zone)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}

	row = db.QueryRow("select * from foo")
	err = row.Scan(&d)
	if err != nil {
		t.Fatal("failed to scan datetime:", err)
	}
	if d.Hour() != 15 || !strings.Contains(d.String(), "JST") {
		t.Fatalf("result should have timezone: %v => %v", zone, d.String())
	}
}

const CurrentTimeStamp = "2006-01-02 15:04:05"

type TimeStamp struct{ *time.Time }

func (t TimeStamp) Scan(value interface{}) error {
	var err error
	switch v := value.(type) {
	case string:
		*t.Time, err = time.Parse(CurrentTimeStamp, v)
	case []byte:
		*t.Time, err = time.Parse(CurrentTimeStamp, string(v))
	default:
		err = errors.New("invalid type for current_timestamp")
	}
	return err
}

func (t TimeStamp) Value() (driver.Value, error) {
	return t.Time.Format(CurrentTimeStamp), nil
}

func TestDateTimeNow(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("failed to open database:", err)
	}
	defer db.Close()

	var d time.Time
	err = db.QueryRow("SELECT datetime('now')").Scan(TimeStamp{&d})
	if err != nil {
		t.Fatal("failed to scan datetime:", err)
	}
}
