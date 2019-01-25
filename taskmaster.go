package main

import (
	"errors"
	"log"
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

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		return err
	}
	pendingTasks, err := db.GetPendingTasksForMinion(minion)
	if err != nil {
		return err
	}
	// split by domain
	pendingForDomain := make(map[uint32][]TaskAssignment)
	for _, pending := range pendingTasks {
		domainID := pending.Task.DomainID
		if pendingForDomain[domainID] == nil {
			pendingForDomain[domainID] = []TaskAssignment{pending}
		} else {
			pendingForDomain[domainID] = append(pendingForDomain[domainID], pending)
		}
	}

	var tasksToAssign []TaskAssignment
	for _, domain := range domains {
		log.Printf("Updating %s\n", domain.Name)

		// Avoid resetting every domain every time we run Update() on the 1st of the month
		if today.Day() == 1 && domain.LastResetDate.Month() != today.Month() {
			db.ResetAllCompletedTasks(domain)
		}

		available, err := db.GetAvailableTasksForDomain(domain)
		if err != nil {
			return err
		}

		// pick random tasks, not in order
		rand.Shuffle(len(available), func(i, j int) {
			available[i], available[j] = available[j], available[i]
		})
		log.Printf("Available: %v\n", available)
		log.Printf("pending: %v\n", pendingForDomain[domain.ID])
		if len(available) == 0 {
			tasksToAssign = append(tasksToAssign, TaskAssignment{Task: NoTask})
			continue // it's nicer to continue than to have another block indented if we used an else
		}

		if len(pendingForDomain[domain.ID]) == 0 {
			// this minion was either added to this Domain, or the Domain is new today or it was reset
			tasksToAssign = append(tasksToAssign, NewTaskAssignment(available[0], minion, today))
			available = available[1:]
			continue
		}

		// Fill any gaps including today with tasks
		additionalTasks, err := fillGapsWithTasks(minion, pendingForDomain[domain.ID], available, today)
		if err != nil {
			return err
		}
		tasksToAssign = append(tasksToAssign, additionalTasks...) // ... is spread

	}

	for _, t := range tasksToAssign {
		if t.Task.ID != NoTask.ID {
			db.SaveTaskAssignment(t)
		}
	}

	return nil
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

// /*
// Split tasks in ones that are assigned to the minion, ones assigned to other people and unassigned (available) ones
// */
// func splitTasks(tasks []Task, minion Minion) ([]Task, []Task, []Task) {
//
// 	var assigned, available, other []Task
//
// 	for _, task := range tasks {
// 		// Note: It might be worth treating any db ID as int64 (despite them bein all unusigned)
// 		if task.AssignedMinionID.Valid {
// 			if task.AssignedMinionID.Int64 == int64(minion.ID) {
// 				assigned = append(assigned, task)
// 			} else {
// 				other = append(other, task)
// 			}
// 		} else {
// 			available = append(available, task)
// 		}
// 	}
//
// 	return assigned, available, other
// }

func newTaskForMinion(domain Domain, minion Minion, date time.Time) (Task, error) {
	var result Task

	return result, nil
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
