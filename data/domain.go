package data

import (
	"time"
)

// Domain is a name for something that has tasks and chores
type Domain struct {
	ID            uint32
	Owner         uint32
	Name          string
	LastResetDate time.Time
}
