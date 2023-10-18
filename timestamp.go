package localfs

import (
	"time"
)

// TODO consider using Carbon library for time stuff

const (
	// default timestamp format
	tsDefaultFormat    = "20060102150405"
	tsLongFormat       = "2006-01-02 15:04:05"
	tsSimpleDateFormat = "2006-1-2"
)

type Timestamp struct {
	time time.Time
}

func (t Timestamp) String() string {
	return t.time.Format(tsDefaultFormat)
}

func (t Timestamp) LongString() string {
	return t.time.Format(tsLongFormat)
}

func (t Timestamp) SimpleDateString() string {
	return t.time.Format(tsSimpleDateFormat)
}

func (t Timestamp) Time() time.Time {
	return t.time
}

func (t Timestamp) SimpleDateAsTime() time.Time {
	return time.Date(t.time.Year(), t.time.Month(), t.time.Day(),
		0, 0, 0, 0, t.time.Location())
}

func NewFromTime(tm time.Time) Timestamp {
	return Timestamp{time: tm}
}

func NewTimestamp(tm string) (Timestamp, error) {
	t, err := time.Parse(tsDefaultFormat, tm)
	if err != nil {
		return Timestamp{}, err
	}
	return Timestamp{t}, nil
}

func NewTimestampSimple(tm string) (Timestamp, error) {
	t, err := time.Parse(tsSimpleDateFormat, tm)
	if err != nil {
		return Timestamp{}, err
	}
	return Timestamp{t}, nil
}

//func FromShortTime(tm time.Time) *Timestamp {
//	return &Timestamp{time: time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, tm.Location())}
//}

//func ToYearString(time time.Time) string {
//	return FromShortTime(time).ShortString()
//}
