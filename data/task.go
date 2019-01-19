package data

import (
	"time"
)

// Task is a chore you do
type Task struct {
	ID           uint32
	DomainID     uint32
	Name         string
	Weekly       bool
	Description  string
	AssignedDate time.Time
}
