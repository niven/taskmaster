package data

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type AssignmentStatus string

const (
	Pending          AssignmentStatus = "pending"
	DoneAndAvailable AssignmentStatus = "done_and_available"
	DoneAndStashed   AssignmentStatus = "done_and_stashed"
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
	ID           uint32
	Task         Task
	MinionID     sql.NullInt64
	AssignedDate pq.NullTime
	AgeInDays    uint32
	Status       AssignmentStatus
}

func NewTaskAssignment(task Task, minion Minion, time time.Time) TaskAssignment {

	result := TaskAssignment{
		Task:         task,
		MinionID:     sql.NullInt64{Int64: int64(minion.ID), Valid: true},
		AssignedDate: pq.NullTime{Time: time, Valid: true},
	}

	return result
}

func TaskFilter(tasks []Task, condition func(t Task) bool) []Task {

	var result []Task
	for _, t := range tasks {
		if condition(t) == true {
			result = append(result, t)
		}
	}

	return result
}
