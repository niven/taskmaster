package logic

import (
	"log"
	"testing"
	"time"

	"github.com/lib/pq"
	. "github.com/niven/taskmaster/data"
	. "github.com/niven/taskmaster/util"
)

func TestAssignTasks(t *testing.T) {

	domains := []Domain{
		Domain{ID: 1},
		Domain{ID: 2},
	}
	availableForDomain := make(map[uint32][]Task)
	availableForDomain[1] = []Task{Task{ID: 999}}
	var assignments []TaskAssignment
	upToIncluding := time.Now() // nothing is assigned, so no gap filling, so just 1 task per domain. The available one and the NoTask

	assignments, err := assignTasks(Minion{ID: 1}, domains, availableForDomain, assignments, upToIncluding)

	if err != nil {
		t.Fail()
	}

	if len(assignments) != 2 {
		t.Fail()
	}

}

func TestAssignTasksForDomainSimple(t *testing.T) {

	minion := Minion{ID: 1}
	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 2)

	available := []Task{
		Task{ID: 123},
	}
	assigned := []TaskAssignment{
		TaskAssignment{
			Task:         Task{ID: 234},
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
		TaskAssignment{
			Task:         Task{ID: 345},
			AssignedDate: pq.NullTime{Valid: true, Time: end},
		},
	}

	additional, err := assignTasksForDomain(minion, available, assigned, end)
	if err != nil {
		t.Fail()
	}
	if len(additional) != 1 {
		t.Fail()
	}
	if additional[0].Task.ID != 123 {
		t.Fail()
	}
}

func TestAssignTasksForDomainNothingAssigned(t *testing.T) {

	minion := Minion{ID: 1}
	available := []Task{
		Task{ID: 123},
	}
	assigned := []TaskAssignment{}

	additional, err := assignTasksForDomain(minion, available, assigned, time.Now())
	if err != nil {
		t.Fail()
	}
	if len(additional) != 1 {
		t.Fail()
	}
	if additional[0].Task.ID != 123 {
		t.Fail()
	}
}

func TestAssignTasksForDomainNothingAvailable(t *testing.T) {

	minion := Minion{ID: 1}
	var available []Task

	additional, err := assignTasksForDomain(minion, available, nil, time.Now())
	if err != nil {
		t.Fail()
	}
	if len(additional) != 1 {
		t.Fail()
	}
	if additional[0].Task.ID != NoTask.ID {
		t.Fail()
	}
}

func TestSplitTaskAssignments(t *testing.T) {

	now := DateFromYYYYMMDD(2019, time.January, 29) // tuesday

	pending := []TaskAssignment{
		// 2 for today
		TaskAssignment{
			AgeInDays:    0,
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 29)},
		},
		TaskAssignment{
			AgeInDays:    0,
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 29)},
		},
		// 8 weekly for each weekday including this one
		TaskAssignment{
			AgeInDays:    0,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 29)}, // tue
		},
		TaskAssignment{
			AgeInDays:    1,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 28)}, // mon
		},
		// these weeklies should be overdue
		TaskAssignment{
			AgeInDays:    2,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 27)}, // sun
		},
		TaskAssignment{
			AgeInDays:    3,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 26)}, // sat
		},
		TaskAssignment{
			AgeInDays:    4,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 25)}, // fri
		},
		TaskAssignment{
			AgeInDays:    5,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 24)}, // thu
		},
		TaskAssignment{
			AgeInDays:    6,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 23)}, // wed
		},
		TaskAssignment{
			AgeInDays:    7,
			Task:         Task{Weekly: true},
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 22)}, // tue
		},
		// a few regular overdue ones
		TaskAssignment{
			AgeInDays:    18,
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 11)},
		},
		TaskAssignment{
			AgeInDays:    21,
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2019, time.January, 8)},
		},
		TaskAssignment{
			AgeInDays:    300,
			AssignedDate: pq.NullTime{Valid: true, Time: DateFromYYYYMMDD(2018, time.December, 2)},
		},
	}

	today, thisWeek, overdue := SplitTaskAssignments(pending, now)
	for _, ta := range today {
		log.Printf("today: %v\n", StrDateFromTime(ta.AssignedDate.Time))
	}
	for _, ta := range thisWeek {
		log.Printf("week: %v\n", StrDateFromTime(ta.AssignedDate.Time))
	}
	for _, ta := range overdue {
		log.Printf("overdue: %v\n", StrDateFromTime(ta.AssignedDate.Time))
	}
	if len(today) != 3 { // 2 for today + 1 weekly for today
		t.Fail()
	}
	if len(thisWeek) != 1 { // only 2 weekly is for this week, the rest is overdue, and onlty 1 is not today
		t.Fail()
	}
	if len(today)+len(thisWeek)+len(overdue) != len(pending) { // the rest is overdue
		t.Fail()
	}
}

func TestFilterTasksExample(t *testing.T) {

	assigned := []TaskAssignment{
		TaskAssignment{
			Task: Task{ID: 123},
		},
		TaskAssignment{
			Task: Task{ID: 123},
		},
		TaskAssignment{
			Task: Task{ID: 234},
		},
		TaskAssignment{
			Task: Task{ID: 456},
		},
	}
	available := []Task{
		Task{ID: 123, Count: 2},
		Task{ID: 234, Count: 3},
		Task{ID: 345, Count: 1},
	}

	result := filterTasks(available, assigned)

	if len(result) != 2 {
		t.Fail()
	}
}

func TestFilterTasks1Remaining(t *testing.T) {

	assigned := []TaskAssignment{
		TaskAssignment{
			Task: Task{ID: 123},
		},
	}
	available := []Task{
		Task{ID: 123, Count: 1},
		Task{ID: 234, Count: 10},
	}

	result := filterTasks(available, assigned)

	if len(result) != 1 {
		t.Fail()
	}
	if result[0].Count != 10 {
		t.Fail()
	}
}

func TestFilterTasksNoOverlap(t *testing.T) {

	assigned := []TaskAssignment{
		TaskAssignment{
			Task: Task{ID: 3},
		},
	}
	available := []Task{
		Task{ID: 1, Count: 10},
		Task{ID: 2, Count: 10},
	}

	result := filterTasks(available, assigned)

	if len(result) != len(available) {
		t.Fail()
	}
	if result[0].Count != 10 || result[1].Count != 10 {
		t.Fail()
	}
}

func TestFilterTasksEmpty(t *testing.T) {

	var assigned []TaskAssignment
	var available []Task

	result := filterTasks(available, assigned)
	if len(result) != 0 {
		t.Fail()
	}
}

func TestFillGapsWithTasksNotEnough(t *testing.T) {

	var minion Minion

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 6)

	assigned := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
	}
	var available []Task
	assigned, err := fillGapsWithTasks(minion, assigned, available, end)
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

func TestFillGapsWithTooManyAssigned(t *testing.T) {

	var minion Minion

	start := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	assigned := []TaskAssignment{
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
		TaskAssignment{
			AssignedDate: pq.NullTime{Valid: true, Time: start},
		},
	}

	var available []Task
	assigned, err := fillGapsWithTasks(minion, assigned, available, start)
	if err == nil {
		t.Fail()
	}
}

func TestFillGapsWithTasksMulti(t *testing.T) {

	var minion Minion

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

	additionalTasks, err := fillGapsWithTasks(minion, assigned, available, end)
	if err != nil {
		t.Fail()
	}
	if len(additionalTasks) != 4 {
		t.Fail()
	}
}

func TestFillGapsWithTasksSingle(t *testing.T) {

	var minion Minion

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

	additionalTasks, err := fillGapsWithTasks(minion, assigned, available, end)

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
