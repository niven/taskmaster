package main

import (
	"database/sql"
	"testing"
	"time"

	. "github.com/niven/taskmaster/data"
)

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
