package util

import (
	"fmt"
	"time"
)

func StrDateFromTime(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
}
