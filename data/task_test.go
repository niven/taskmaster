package data

import (
	"testing"
)

func TestTaskAssignmentFilter(t *testing.T) {

	empty := TaskAssignmentFilter([]TaskAssignment{}, func(t TaskAssignment) bool { return t.AgeInDays > 0 })
	if len(empty) != 0 {
		t.Fail()
	}

	single := TaskAssignmentFilter([]TaskAssignment{TaskAssignment{AgeInDays: 0}, TaskAssignment{AgeInDays: 1}}, func(t TaskAssignment) bool { return t.AgeInDays > 0 })
	if len(single) != 1 {
		t.Fail()
	}

	both := TaskAssignmentFilter([]TaskAssignment{TaskAssignment{AgeInDays: 1}, TaskAssignment{AgeInDays: 1}}, func(t TaskAssignment) bool { return t.AgeInDays > 0 })
	if len(both) != 2 {
		t.Fail()
	}
}

func TestTaskFilter(t *testing.T) {

	empty := TaskFilter([]Task{}, func(t Task) bool { return t.Weekly })
	if len(empty) != 0 {
		t.Fail()
	}

	single := TaskFilter([]Task{Task{Weekly: true}, Task{Weekly: false}}, func(t Task) bool { return t.Weekly })
	if len(single) != 1 {
		t.Fail()
	}

	both := TaskFilter([]Task{Task{Weekly: true}, Task{Weekly: true}}, func(t Task) bool { return t.Weekly })
	if len(both) != 2 {
		t.Fail()
	}
}
