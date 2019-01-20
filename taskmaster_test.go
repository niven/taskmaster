package main

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/lib/pq"
	. "github.com/niven/taskmaster/data"
)

func TestFillGapsWithTasksNotEnough(t *testing.T) {

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6)

	assigned := []Task{
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
	}
	var available []Task
	_, err := fillGapsWithTasks(assigned, available, end)
	if err == nil {
		t.Fail()
	}

}

func TestFillGapsWithTasksMulti(t *testing.T) {

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6)

	assigned := []Task{
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: start.AddDate(0, 0, 2)},
		},
		Task{
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
	assigned := []Task{
		Task{
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

	tasks := []Task{
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2008, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2007, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2006, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
	}

	oldest, err := findOldestTaskTime(tasks)
	if err != nil {
		t.Fail()
	}

	if !oldest.Equal(tasks[3].AssignedDate.Time) {
		t.Fail()
	}
}

func TestFindOldestTaskTimeEmpty(t *testing.T) {

	var tasks []Task

	_, err := findOldestTaskTime(tasks)
	if err == nil {
		t.Fail()
	}
}

func TestFindOldestTaskTimeSingle(t *testing.T) {

	tasks := []Task{
		Task{
			AssignedDate: pq.NullTime{Valid: true, Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
		},
	}

	oldest, err := findOldestTaskTime(tasks)
	if err != nil {
		t.Fail()
	}
	if !oldest.Equal(tasks[0].AssignedDate.Time) {
		t.Fail()
	}
}

func TestSplitTasksMany(t *testing.T) {

	m := Minion{
		ID: 1,
	}
	tasks := []Task{
		Task{
			AssignedMinionID: sql.NullInt64{Valid: true, Int64: 1},
		},
		Task{
			AssignedMinionID: sql.NullInt64{Valid: true, Int64: 1},
		},
		Task{
			AssignedMinionID: sql.NullInt64{Valid: true, Int64: 2},
		},
		Task{},
	}
	assigned, available, other := splitTasks(tasks, m)

	if len(assigned)+len(available)+len(other) != len(tasks) {
		t.Fail()
	}
	if len(assigned) != 2 {
		t.Fail()
	}
	if len(available) != 1 {
		t.Fail()
	}
	if len(other) != 1 {
		t.Fail()
	}

}

func TestSplitTasksSingle(t *testing.T) {

	m := Minion{
		ID: 1,
	}
	tasks := []Task{
		Task{
			AssignedMinionID: sql.NullInt64{Valid: true, Int64: 1},
		},
	}
	assigned, available, other := splitTasks(tasks, m)

	if len(assigned)+len(available)+len(other) != len(tasks) {
		t.Fail()
	}

	if len(available)+len(other) != 0 {
		t.Fail()
	}
	if len(assigned) != 1 {
		t.Fail()
	}

}

func TestSplitTasksEmpty(t *testing.T) {

	m := Minion{
		ID: 1,
	}
	var tasks []Task
	assigned, available, other := splitTasks(tasks, m)

	if len(assigned)+len(available)+len(other) != len(tasks) {
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

func TestStrDateFromTime(t *testing.T) {
	from := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	strDate := strDateFromTime(from)
	if strDate != "2009-11-10" {
		t.Fail()
	}
}
