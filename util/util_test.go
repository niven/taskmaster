package util

import (
	"testing"
	"time"
)

func TestDateEqual(t *testing.T) {
	if !DateEqual(DateFromYYYYMMDD(2017, 3, 12), DateFromYYYYMMDD(2017, 3, 12)) {
		t.Fail()
	}

	if !DateEqual(DateFromYYYYMMDD(2017, 3, 12), time.Date(2017, 3, 12, 23, 11, 16, 0, time.UTC)) {
		t.Fail()
	}
}

func TestDateFromYYYYMMDD(t *testing.T) {

	if StrDateFromTime(DateFromYYYYMMDD(2017, 3, 12)) != "2017-03-12" {
		t.Fail()
	}

	if StrDateFromTime(DateFromYYYYMMDD(2000, 1, 1)) != "2000-01-01" {
		t.Fail()
	}

	if StrDateFromTime(DateFromYYYYMMDD(1999, 12, 31)) != "1999-12-31" {
		t.Fail()
	}

	if StrDateFromTime(DateFromYYYYMMDD(2020, 2, 22)) != "2020-02-22" {
		t.Fail()
	}

}

func TestIsWeekDay(t *testing.T) {

	weekdays := !IsWeekendDay(time.Monday) && !IsWeekendDay(time.Tuesday) && !IsWeekendDay(time.Wednesday) && !IsWeekendDay(time.Thursday) && !IsWeekendDay(time.Friday)
	if !weekdays {
		t.Fail()
	}

	weekendDays := IsWeekendDay(time.Saturday) && IsWeekendDay(time.Sunday)
	if !weekendDays {
		t.Fail()
	}
}

func TestStrDateFromTime(t *testing.T) {
	from := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	strDate := StrDateFromTime(from)
	if strDate != "2009-11-10" {
		t.Fail()
	}
}
