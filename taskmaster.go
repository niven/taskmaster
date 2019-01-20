package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	. "github.com/niven/taskmaster/data"
	"github.com/niven/taskmaster/db"
)

/*
	Simple case: get 1 new task per Domain for today
	More complicated: find all past days that have no tasks assigned yet, and assign tasks for them
	Most complicated case: a Domain is through all tasks for the month. Remove all completed ones and issue more
*/
func Update(minion Minion) error {

	domains, err := db.GetDomainsForMinion(minion)
	if err != nil {
		return err
	}

	today := time.Now()
	var tasksToAssign []Task
	for _, domain := range domains {
		log.Printf("Updating %s\n", domain.Name)
		tasks, err := db.GetAllTasks(domain)
		if err != nil {
			return err
		}

		// assigned, available, other := splitTasks(tasks, minion)

		var assignedTasks []Task
		for _, task := range tasks {
			// Note: It might be worth treating any db ID as int64 (despite them bein all unusigned)
			if task.AssignedMinionID.Valid && task.AssignedMinionID.Int64 == int64(minion.ID) {
				assignedTasks = append(assignedTasks, task)
			}
		}

		if len(assignedTasks) == 0 {
			// cool, just assign a new task.
			// this minion was either added to this Domain, or the Domain is new today
			newTask, err := newTaskForMinion(domain, minion, today)
			if err != nil {
				return err
			}
			tasksToAssign = append(tasksToAssign, newTask)
			// it's nicer to continue than to have another block indented if we used an else
			continue
		}

		// find the oldest task, and calculate the number of tasks between that day and today
		var oldest time.Time = assignedTasks[0].AssignedDate.Time
		for _, task := range assignedTasks {
			if task.AssignedDate.Time.Before(oldest) {
				oldest = task.AssignedDate.Time
			}
		}
		dates := makeContiguousDates(oldest, today)

		if len(dates) > len(tasks) {
			return errors.New("Somehow there are more tasks than days...")
		}

		// fill the gaps, days someone didn't log in still generate tasks
		if len(dates) < len(assignedTasks) {
			// make a map so we can easily find the missing dates
			// and use strDates so it's always YYYY-MM-DD and not some time object with a milli off
			tasksByDate := make(map[string]Task)
			for _, task := range tasks {
				tasksByDate[strDateFromTime(task.AssignedDate.Time)] = task
			}
			for _, date := range dates {
				if _, exists := tasksByDate[strDateFromTime(date)]; !exists {
					newTask, err := newTaskForMinion(domain, minion, date)
					if err != nil {
						return err
					}
					tasksToAssign = append(tasksToAssign, newTask)
				}
			}
		}

	}

	// save all tasks

	return nil
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

func strDateFromTime(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
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

	endStr := strDateFromTime(d2)
	for strings.Compare(strDateFromTime(currentTime), endStr) == -1 {
		currentTime = currentTime.AddDate(0, 0, 1)
		result = append(result, currentTime)
	}

	return result
}
