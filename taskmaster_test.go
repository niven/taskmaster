package main

import (
	"testing"
	"time"
)

func TestMakeContiguousDatesSame(t *testing.T) {
	from := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	to := from

	dates := makeContiguousDates(from, to)
	if len(dates) != 1 {
		t.Fail()
	}

	if !dates[0].Equal(from) {
		t.Fail()
	}
}

func TestMakeContiguousDatesOrder(t *testing.T) {
	from := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	to := time.Date(2009, time.November, 8, 23, 0, 0, 0, time.UTC)

	dates := makeContiguousDates(from, to)
	if len(dates) != 3 {
		t.Fail()
	}
}

func TestStrDateFromTime(t *testing.T) {
	from := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	strDate := strDateFromTime(from)
	if strDate != "2009-11-10" {
		t.Fail()
	}
}
