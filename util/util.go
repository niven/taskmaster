package util

import (
	"fmt"
	"time"
)

// might depend on locale
func IsWeekendDay(day time.Weekday) bool {
	return day == time.Saturday || day == time.Sunday
}

func DateFromYYYYMMDD(yyyy int, mm time.Month, dd int) time.Time {
	return time.Date(yyyy, time.January, dd, 0, 0, 0, 0, time.UTC)
}

func StrDateFromTime(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
}
