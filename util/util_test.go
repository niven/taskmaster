package util

import (
	"testing"
	"time"
)

func TestStrDateFromTime(t *testing.T) {
	from := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	strDate := StrDateFromTime(from)
	if strDate != "2009-11-10" {
		t.Fail()
	}
}
