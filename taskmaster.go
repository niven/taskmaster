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
	if today.Day() == 1 {
		db.ResetAllCompletedTasks()
	}

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		return err
	}

	var tasksToAssign []Task
	for _, domain := range domains {
		log.Printf("Updating %s\n", domain.Name)

		tasks, err := db.GetAllTasks(domain)
		if err != nil {
			return err
		}

		assigned, available, _ := splitTasks(tasks, minion)

		if len(available) == 0 {
			tasksToAssign = append(tasksToAssign, NoTask)
			continue // it's nicer to continue than to have another block indented if we used an else
		}

		if len(assigned) == 0 {
			// this minion was either added to this Domain, or the Domain is new today or it was reset
			tasksToAssign = append(tasksToAssign, available[0])
			available = available[1:]
			continue
		}

		// Fill any gaps including today with tasks
		additionalTasks, err := fillGapsWithTasks(assigned, available, today)
		if err != nil {
			return err
		}
		tasksToAssign = append(tasksToAssign, additionalTasks...) // ... is spread

	}

	for _, t := range tasksToAssign {
		if t.ID != NoTask.ID {
			db.SaveTaskAssignment(t.ID, minion.ID, t.AssignedDate.Time)
		}
	}

	return nil
}

/*
	For every day that doesn't have an assigned task, pick one from the available ones
*/
func fillGapsWithTasks(assigned, available []Task, upToIncluding time.Time) ([]Task, error) {

	var result []Task

	// find the oldest task, and calculate the number of tasks between that day and today
	oldest, _ := findOldestTaskTime(assigned)
	// ignoring err since that only happens when assigned has no elements

	dates := makeContiguousDates(oldest, upToIncluding)

	if len(assigned) > len(dates) {
		return result, errors.New("Somehow there are more tasks assigned than days...")
	}

	// fill the gaps, days someone didn't log in still generate tasks
	if len(assigned) < len(dates) {

		// TODO: Remove assigned tasks to avoid dupes

		// pick random tasks, not in order
		rand.Shuffle(len(available), func(i, j int) {
			available[i], available[j] = available[j], available[i]
		})

		// make a map so we can easily find the missing dates
		// and use strDates so it's always YYYY-MM-DD and not some time object with a milli off
		tasksByDate := make(map[string]Task)
		for _, task := range assigned {
			tasksByDate[util.StrDateFromTime(task.AssignedDate.Time)] = task
		}

		for _, date := range dates {
			if _, exists := tasksByDate[util.StrDateFromTime(date)]; !exists {

				if len(available) == 0 {
					result = append(result, NoTask)
				} else {
					result = append(result, available[0])
					available = available[1:]
				}

			}
		}
	}

	return result, nil
}

func findOldestTaskTime(tasks []Task) (time.Time, error) {

	var oldest time.Time
	if len(tasks) == 0 {
		return oldest, errors.New("Empty task list")
	}

	oldest = tasks[0].AssignedDate.Time
	for _, task := range tasks {
		if task.AssignedDate.Time.Before(oldest) {
			oldest = task.AssignedDate.Time
		}
	}
	return oldest, nil
}

/*
Split tasks in ones that are assigned to the minion, ones assigned to other people and unassigned (available) ones
*/
func splitTasks(tasks []Task, minion Minion) ([]Task, []Task, []Task) {

	var assigned, available, other []Task

	for _, task := range tasks {
		// Note: It might be worth treating any db ID as int64 (despite them bein all unusigned)
		if task.AssignedMinionID.Valid {
			if task.AssignedMinionID.Int64 == int64(minion.ID) {
				assigned = append(assigned, task)
			} else {
				other = append(other, task)
			}
		} else {
			available = append(available, task)
		}
	}

	return assigned, available, other
}

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
