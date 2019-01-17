package taskmaster

import (
	"testing"
)

func TestGetAllTasks(t *testing.T) {

	tasks := getAllTasks(0)

	if len(tasks) != 3 {
		t.Error("Expected number of tasks to be 3")
	}

}
