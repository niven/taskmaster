package data

import (
	"database/sql"
	"time"

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
	ID            uint32
	Task          Task
	MinionID      sql.NullInt64
	AssignedDate  pq.NullTime
	CompletedDate pq.NullTime
}

func NewTaskAssignment(task Task, minion Minion, time time.Time) TaskAssignment {

	result := TaskAssignment{
		Task:         task,
		MinionID:     sql.NullInt64{Int64: int64(minion.ID), Valid: true},
		AssignedDate: pq.NullTime{Time: time, Valid: true},
	}

	return result
}
