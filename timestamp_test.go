package versionfs

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const defaultTS = "20221019140203"

func TestTimestamp_NewFromTime(t *testing.T) {
	t.Parallel()
	date := time.Date(2022, time.October, 19, 14, 2, 3, 0, time.UTC)
	ts := NewFromTime(date)
	assert.Equal(t, defaultTS, ts.String())
	assert.Equal(t, "2022-10-19", ts.SimpleDateString())
	assert.Equal(t, "2022-10-19 14:02:03", ts.LongString())
	tm := ts.Time()
	assert.Equal(t, 2022, tm.Year())
	assert.Equal(t, time.October, tm.Month())
	assert.Equal(t, 19, tm.Day())
	assert.Equal(t, 14, tm.Hour())
	assert.Equal(t, 2, tm.Minute())
	assert.Equal(t, 3, tm.Second())
	assert.Equal(t, 0, tm.Nanosecond())
	tm = ts.SimpleDateAsTime()
	assert.Equal(t, 2022, tm.Year())
	assert.Equal(t, time.October, tm.Month())
	assert.Equal(t, 19, tm.Day())
	assert.Equal(t, 0, tm.Hour())
	assert.Equal(t, 0, tm.Minute())
	assert.Equal(t, 0, tm.Second())
	assert.Equal(t, 0, tm.Nanosecond())
}

func TestTimestamp_New(t *testing.T) {
	t.Parallel()
	ts, err := NewTimestamp(defaultTS)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, defaultTS, ts.String())
	assert.Equal(t, "2022-10-19", ts.SimpleDateString())
	assert.Equal(t, "2022-10-19 14:02:03", ts.LongString())
	tm := ts.Time()
	assert.Equal(t, 2022, tm.Year())
	assert.Equal(t, time.October, tm.Month())
	assert.Equal(t, 19, tm.Day())
	assert.Equal(t, 14, tm.Hour())
	assert.Equal(t, 2, tm.Minute())
	assert.Equal(t, 3, tm.Second())
	assert.Equal(t, 0, tm.Nanosecond())
	tm = ts.SimpleDateAsTime()
	assert.Equal(t, 2022, tm.Year())
	assert.Equal(t, time.October, tm.Month())
	assert.Equal(t, 19, tm.Day())
	assert.Equal(t, 0, tm.Hour())
	assert.Equal(t, 0, tm.Minute())
	assert.Equal(t, 0, tm.Second())
	assert.Equal(t, 0, tm.Nanosecond())
}

func TestTimestamp_NewErr(t *testing.T) {
	t.Parallel()
	_, err := NewTimestamp("foo")
	assert.Equal(t, "parsing time \"foo\" as \"20060102150405\": cannot parse \"foo\" as \"2006\"", err.Error())
}

func TestTimestamp_NewSimple(t *testing.T) {
	t.Parallel()
	ts, err := NewTimestampSimple("2022-10-19")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "2022-10-19", ts.SimpleDateString())
	tm := ts.Time()
	assert.Equal(t, 2022, tm.Year())
	assert.Equal(t, time.October, tm.Month())
	assert.Equal(t, 19, tm.Day())
	assert.Equal(t, 0, tm.Hour())
	assert.Equal(t, 0, tm.Minute())
	assert.Equal(t, 0, tm.Second())
}

func TestTimestamp_NewSimpleErr(t *testing.T) {
	t.Parallel()
	_, err := NewTimestampSimple("invalid")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parsing time")
}

//func TestNewReadableTSFromShortTime(t *testing.T) {
//	t.Parallel()
//	date := time.Date(2022, 1, 9, 1, 2, 3, 0, time.UTC)
//	ts := FromShortTime(date)
//	assert.Equal(t, "20220109000000", ts.String())
//	assert.Equal(t, "2022-1-9", ts.ShortString())
//	assert.Equal(t, "2022-01-09 00:00:00", ts.LongString())
//	tm := ts.Time()
//	assert.Equal(t, 2022, tm.Year())
//	assert.Equal(t, time.January, tm.Month())
//	assert.Equal(t, 9, tm.Day())
//	assert.Equal(t, 0, tm.Hour())
//	assert.Equal(t, 0, tm.Minute())
//	assert.Equal(t, 0, tm.Second())
//	assert.Equal(t, 0, tm.Nanosecond())
//	tm = ts.ShortTime()
//	assert.Equal(t, 2022, tm.Year())
//	assert.Equal(t, time.January, tm.Month())
//	assert.Equal(t, 9, tm.Day())
//	assert.Equal(t, 0, tm.Hour())
//	assert.Equal(t, 0, tm.Minute())
//	assert.Equal(t, 0, tm.Second())
//	assert.Equal(t, 0, tm.Nanosecond())
//}
//
//func TestReadableTSFromShortString(t *testing.T) {
//	t.Parallel()
//	const s = "2022-1-9"
//	ts, err := FromShortString(s)
//	if err != nil {
//		t.Fatal(err)
//	}
//	assert.Equal(t, "2022-1-9", ts.ShortString())
//	assert.Equal(t, "2022-01-09 00:00:00", ts.LongString())
//	tm := ts.Time()
//	assert.Equal(t, 2022, tm.Year())
//	assert.Equal(t, time.January, tm.Month())
//	assert.Equal(t, 9, tm.Day())
//	assert.Equal(t, 0, tm.Hour())
//	assert.Equal(t, 0, tm.Minute())
//	assert.Equal(t, 0, tm.Second())
//	assert.Equal(t, 0, tm.Nanosecond())
//}
//
//func TestNewReadableTSFromShortStringErr(t *testing.T) {
//	t.Parallel()
//	ts, err := FromShortString("foo")
//	assert.Nil(t, ts)
//	assert.Equal(t, "parsing time \"foo\" as \"2006-1-2\": cannot parse \"foo\" as \"2006\"", err.Error())
//}
//
//func TestToYearString(t *testing.T) {
//	t.Parallel()
//	date := time.Date(2022, 1, 9, 1, 2, 3, 0, time.UTC)
//	assert.Equal(t, "2022-1-9", ToYearString(date))
//}
