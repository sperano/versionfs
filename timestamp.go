package versionfs

import (
	"time"
)

const (
	// tsDefaultFormat is the timestamp format used in filenames: YYYYMMDDHHmmss
	tsDefaultFormat = "20060102150405"
	// tsLongFormat is a human-readable timestamp format: YYYY-MM-DD HH:mm:ss
	tsLongFormat = "2006-01-02 15:04:05"
	// tsSimpleDateFormat is a simple date format: YYYY-M-D
	tsSimpleDateFormat = "2006-1-2"
)

// Timestamp represents a point in time used for file versioning.
// It wraps a time.Time and provides multiple formatting options.
type Timestamp struct {
	time time.Time
}

// String returns the timestamp in the default format (YYYYMMDDHHmmss).
// This format is used in filenames.
//
// Example: "20231019140523"
func (t Timestamp) String() string {
	return t.time.Format(tsDefaultFormat)
}

// LongString returns the timestamp in a human-readable format.
//
// Example: "2023-10-19 14:05:23"
func (t Timestamp) LongString() string {
	return t.time.Format(tsLongFormat)
}

// SimpleDateString returns the timestamp as a simple date string.
//
// Example: "2023-10-19"
func (t Timestamp) SimpleDateString() string {
	return t.time.Format(tsSimpleDateFormat)
}

// Time returns the underlying time.Time value.
func (t Timestamp) Time() time.Time {
	return t.time
}

// SimpleDateAsTime returns a time.Time with the date components but time set to midnight.
// Useful for date-only comparisons.
func (t Timestamp) SimpleDateAsTime() time.Time {
	return time.Date(t.time.Year(), t.time.Month(), t.time.Day(),
		0, 0, 0, 0, t.time.Location())
}

// NewFromTime creates a Timestamp from a time.Time value.
//
// Example:
//
//	ts := localfs.NewFromTime(time.Now())
func NewFromTime(tm time.Time) Timestamp {
	return Timestamp{time: tm}
}

// NewTimestamp parses a timestamp string in the default format (YYYYMMDDHHmmss).
// Returns an error if the string cannot be parsed.
//
// Example:
//
//	ts, err := localfs.NewTimestamp("20231019140523")
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewTimestamp(tm string) (Timestamp, error) {
	t, err := time.Parse(tsDefaultFormat, tm)
	if err != nil {
		return Timestamp{}, err
	}
	return Timestamp{t}, nil
}

// NewTimestampSimple parses a timestamp string in simple date format (YYYY-M-D).
// Returns an error if the string cannot be parsed.
// The time component is set to midnight.
//
// Example:
//
//	ts, err := localfs.NewTimestampSimple("2023-10-19")
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewTimestampSimple(tm string) (Timestamp, error) {
	t, err := time.Parse(tsSimpleDateFormat, tm)
	if err != nil {
		return Timestamp{}, err
	}
	return Timestamp{t}, nil
}
