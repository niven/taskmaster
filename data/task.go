package data

import (
	"database/sql"

	"github.com/lib/pq"
)

// Task is a chore you do
type Task struct {
	ID               uint32
	DomainID         uint32
	Name             string
	Weekly           bool
	Description      sql.NullString
	AssignedMinionID sql.NullInt64
	AssignedDate     pq.NullTime
	CompletedDate    pq.NullTime
}
