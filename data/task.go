package data

import (
	"database/sql"

	"github.com/lib/pq"
)

// Task is a chore you do
type Task struct {
	ID          uint32
	DomainID    uint32
	Name        string
	Weekly      bool
	Count       uint32
	Description sql.NullString
}

var NoTask = Task{
	ID:   0,
	Name: "Nothing to do",
}

type TaskAssignment struct {
	Task          Task
	MinionID      sql.NullInt64
	AssignedDate  pq.NullTime
	CompletedDate pq.NullTime
}
