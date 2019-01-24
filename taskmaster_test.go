package main

import (
	"log"
	"testing"
	"time"

	"github.com/lib/pq"
	. "github.com/niven/taskmaster/data"
)

func TestFillGapsWithTasksNotEnough(t *testing.T) {

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6)

	assigned := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
	}
	var available []Task
	assigned, err := fillGapsWithTasks(assigned, available, end)
	if err != nil {
		t.Fail()
	}
	noTaskCount := 0
	for _, t := range assigned {
		if t.Task.ID == NoTask.ID {
			noTaskCount++
		}
	}
	if noTaskCount != 6 {
		t.Fail()
	}
}

func TestFillGapsWithTasksMulti(t *testing.T) {

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6)

	assigned := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start.AddDate(0, 0, 2)},
		},
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start.AddDate(0, 0, 4)},
		},
	}
	available := []Task{
		Task{},
		Task{},
		Task{},
		Task{},
		Task{},
		Task{},
		Task{},
	}

	additionalTasks, err := fillGapsWithTasks(assigned, available, end)
	if err != nil {
		t.Fail()
	}
	if len(additionalTasks) != 4 {
		t.Fail()
	}
}

func TestFillGapsWithTasksSingle(t *testing.T) {

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	log.Printf("%v %v", start, end)
	assigned := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
	}
	available := []Task{
		Task{},
	}

	additionalTasks, err := fillGapsWithTasks(assigned, available, end)

	if err != nil {
		t.Fail()
	}
	if len(additionalTasks) != 1 {
		t.Fail()
	}

}

func TestFindOldestTaskTimeMulti(t *testing.T) {

	tasks := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2008, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2007, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2006, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
	}

	oldest, err := findOldestAssignmentTime(tasks)
	if err != nil {
		t.Fail()
	}

	if !oldest.Equal(tasks[3].AssignedDate.Time) {
		t.Fail()
	}
}

func TestFindOldestTaskTimeEmpty(t *testing.T) {

	var assignments []TaskAssignment

	_, err := findOldestAssignmentTime(assignments)
	if err == nil {
		t.Fail()
	}
}

func TestFindOldestTaskTimeSingle(t *testing.T) {

	assignments := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
	}

	oldest, err := findOldestAssignmentTime(assignments)
	if err != nil {
		t.Fail()
	}
	if !oldest.Equal(assignments[0].AssignedDate.Time) {
		t.Fail()
	}
}

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
