package logic

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	. "github.com/niven/taskmaster/data"
	"github.com/niven/taskmaster/db"
	"github.com/niven/taskmaster/util"
)

/*
	So this is a bit too complex at the moment:
	- Every day a minion gets a new task for each of their Domains (I might rename that)
	- At the 1st of the month, every task gets 'shuffled back in'
	- After completing a task, the minion can stash the card or shuffle back in
	- If all cards run out before the end of the month.... that's fine I guess?
	- If a minion has missed days, retroactively assign tasks to them
*/
func Update(minion Minion) error {

	today := time.Now()

	domains := db.GetDomainsForMinion(minion)

	assignments := db.AssignmentRetrieveForMinion(minion, true)

	availableForDomain := make(map[uint32][]Task)

	for _, domain := range domains {

		// Avoid resetting every domain every time we run Update() on the 1st of the month
		if today.Day() == 1 && domain.LastResetDate.Month() != today.Month() {
			db.ResetAllCompletedTasks(domain)
		}

		available, err := db.GetAvailableTasksForDomain(domain)
		if err != nil {
			return err
		}
		availableForDomain[domain.ID] = available
	}

	tasksToAssign, err := assignTasks(minion, domains, availableForDomain, assignments, today)
	if err != nil {
		return err
	}

	for _, t := range tasksToAssign {
		if t.Task.ID != NoTask.ID {
			db.AssignmentInsert(t)
		}
	}

	return nil
}

func assignTasks(minion Minion, domains []Domain, availableForDomain map[uint32][]Task, assignments []TaskAssignment, upToIncluding time.Time) ([]TaskAssignment, error) {

	var result []TaskAssignment

	for _, domain := range domains {

		available := availableForDomain[domain.ID]

		// for convenience, keep only the ones for this domain
		assignments := TaskAssignmentFilter(assignments, func(ta TaskAssignment) bool { return ta.Task.DomainID == domain.ID })

		// filter out tasks we alread have pending. No need to get laundry assigned after having overdue laundry
		available = TaskFilter(available, func(t Task) bool {
			for _, assigment := range assignments {
				if assigment.Task.ID == t.ID {
					return false
				}
			}
			return true
		})

		additional, err := assignTasksForDomain(minion, available, assignments, upToIncluding)
		if err != nil {
			return nil, err
		}
		result = append(result, additional...) // ... is spread
	}

	return result, nil
}

func assignTasksForDomain(minion Minion, available []Task, assignments []TaskAssignment, upToIncluding time.Time) ([]TaskAssignment, error) {

	var result []TaskAssignment

	if len(available) == 0 {
		return []TaskAssignment{TaskAssignment{Task: NoTask}}, nil
	}

	// pick random tasks, not in order
	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	if len(assignments) == 0 {
		// this minion was either added to this Domain, or the Domain is new today or it was reset
		// result = append(result, NewTaskAssignment(available[0], minion, upToIncluding))
		return []TaskAssignment{NewTaskAssignment(available[0], minion, upToIncluding)}, nil
	}

	// Fill any gaps including today with tasks
	result, err := fillGapsWithTasks(minion, assignments, available, upToIncluding)
	return result, err

}

// split pending assignments into 3 lists: those for the current date, overdue ones and weekly ones
// Note: weekly ones become overdue based on the day the week starts
func SplitTaskAssignments(pendingTaskAssignments []TaskAssignment, now time.Time) ([]TaskAssignment, []TaskAssignment, []TaskAssignment) {

	var today, thisWeek, overdue []TaskAssignment

	for _, assignment := range pendingTaskAssignments {

		if assignment.AgeInDays == 0 {
			today = append(today, assignment)
			continue
		}
		if !assignment.Task.Weekly || assignment.AgeInDays > 6 {
			overdue = append(overdue, assignment)
			continue
		}

		// previous weekend or any weekday later than today must belong to last week
		day := assignment.AssignedDate.Time.Weekday()
		if day > now.Weekday() || util.IsWeekendDay(day) {
			overdue = append(overdue, assignment)
		} else {
			thisWeek = append(thisWeek, assignment)
		}
	}

	return today, thisWeek, overdue
}

/*
	For every day that doesn't have an assigned task, pick one from the available ones
*/
func fillGapsWithTasks(minion Minion, assigned []TaskAssignment, available []Task, upToIncluding time.Time) ([]TaskAssignment, error) {

	var result []TaskAssignment

	// find the oldest task, and calculate the number of tasks between that day and today
	oldest, _ := findOldestAssignmentTime(assigned)
	// ignoring err since that only happens when assigned has no elements

	dates := makeContiguousDates(oldest, upToIncluding)

	if len(assigned) > len(dates) {
		return result, errors.New("Somehow there are more tasks assigned than days...")
	}

	// fill the gaps, days someone didn't log in still generate tasks
	if len(assigned) < len(dates) {

		// TODO: Remove assigned tasks to avoid dupes

		// make a map so we can easily find the missing dates
		// and use strDates so it's always YYYY-MM-DD and not some time object with a milli off
		tasksByDate := make(map[string]TaskAssignment)
		for _, task := range assigned {
			tasksByDate[util.StrDateFromTime(task.AssignedDate.Time)] = task
		}

		for _, date := range dates {
			if _, exists := tasksByDate[util.StrDateFromTime(date)]; !exists {

				if len(available) == 0 {
					result = append(result, TaskAssignment{Task: NoTask})
				} else {
					result = append(result, NewTaskAssignment(available[0], minion, date))
					available = available[1:]
				}

			}
		}
	}

	return result, nil
}

func findOldestAssignmentTime(assignments []TaskAssignment) (time.Time, error) {

	var oldest time.Time
	if len(assignments) == 0 {
		return oldest, errors.New("Empty assignments list")
	}

	oldest = assignments[0].AssignedDate.Time
	for _, assignment := range assignments {
		if assignment.AssignedDate.Time.Before(oldest) {
			oldest = assignment.AssignedDate.Time
		}
	}
	return oldest, nil
}

// fill the interval [d1, d2] of dates
func makeContiguousDates(d1, d2 time.Time) []time.Time {
	// Relevant reading: https://infiniteundo.com/post/25326999628/falsehoods-programmers-believe-about-time
	// increment the oldest time by 1 day until we reach todays date

	if d1.After(d2) {
		d1, d2 = d2, d1
	}

	var result = []time.Time{d1}

	currentTime := d1

	endStr := util.StrDateFromTime(d2)
	for strings.Compare(util.StrDateFromTime(currentTime), endStr) == -1 {
		currentTime = currentTime.AddDate(0, 0, 1)
		result = append(result, currentTime)
	}

	return result
}
