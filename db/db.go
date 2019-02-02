package db

import (
	"database/sql"
	"errors"
	"log"

	// have to import with an underbar alias since we need the init() to run
	_ "github.com/lib/pq"

	"github.com/niven/taskmaster/config"
	. "github.com/niven/taskmaster/data"
	"github.com/niven/taskmaster/util"
)

var (
	db *sql.DB
)

func init() {

	config.ReadEnvironmentVars()

	// if we do db, err := foo() then this 'db' shadows the global one
	var err error
	db, err = sql.Open("postgres", config.EnvironmentVars["DATABASE_URL"])

	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
}

func CreateMinion(email, name string) error {

	_, err := db.Exec("INSERT INTO minions (email, name) VALUES($1,$2)", email, name)

	if err != nil {
		log.Printf("Error inserting new minion: %q", err)
		return err
	}

	return nil
}

func LoadMinion(email string, m *Minion) bool {

	row := db.QueryRow("SELECT * FROM minions WHERE email = $1", email)

	err := row.Scan(&m.ID, &m.Email, &m.Name)
	if err != nil && err == sql.ErrNoRows {
		return false
	}

	return true
}

func CreateNewDomain(minion Minion, domainName string) error {
	_, err := db.Exec("INSERT INTO domains (owner, name) VALUES($1, $2)", minion.ID, domainName)

	if err != nil {
		log.Printf("Error inserting new domain: %q", err)
		return err
	}

	return nil
}

func CreateNewTask(task Task) error {
	_, err := db.Exec("INSERT INTO tasks (domain_id, name, weekly, count) VALUES($1, $2, $3, $4)", task.DomainID, task.Name, task.Weekly, task.Count)

	if err != nil {
		log.Printf("Error inserting new task: %q", err)
		return err
	}

	return nil
}

func GetDomainByID(domainID uint32) (Domain, error) {

	row := db.QueryRow("SELECT id, owner, name, last_reset_date FROM domains WHERE id = $1", domainID)

	var result Domain

	err := row.Scan(&result.ID, &result.Owner, &result.Name, &result.LastResetDate)
	if err == sql.ErrNoRows {
		return result, err
	}

	return result, nil
}

func GetDomainsForMinion(m Minion) ([]Domain, error) {

	rows, err := db.Query("SELECT d.id, d.owner, d.name, d.last_reset_date, COUNT(t.id) AS task_count FROM domains d LEFT JOIN tasks t ON d.id = t.domain_id WHERE owner = $1 GROUP BY d.id", m.ID)

	if err != nil {
		return nil, err
	}

	var result []Domain
	defer rows.Close()
	for rows.Next() {
		var d Domain

		if err := rows.Scan(&d.ID, &d.Owner, &d.Name, &d.LastResetDate, &d.TaskCount); err != nil {
			log.Printf("Error scanning domains: %q", err)
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
}

func ReadAllMinions() ([]Minion, error) {

	rows, err := db.Query("SELECT * FROM minions")
	if err != nil {
		log.Printf("Error reading minions: %q", err)
		return nil, err
	}

	result := []Minion{}

	defer rows.Close()
	for rows.Next() {
		var m Minion

		if err := rows.Scan(&m.ID, &m.Email, &m.Name); err != nil {
			log.Printf("Error scanning minion: %q", err)
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
}

func ResetAllCompletedTasks(domain Domain) error {

	result, err := db.Exec("DELETE FROM task_assignments WHERE completed_on IS NOT NULL AND domain_id = $1", domain.ID)
	if err != nil {
		return err
	}
	result, err = db.Exec("UPDATE domains SET last_reset_date = CURRENT_DATE WHERE id = $1", domain.ID)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err == nil {
		log.Printf("Reset %d tasks\n", count)
	} else {
		return err
	}
	return nil
}

func GetAvailableTasksForDomain(domain Domain) ([]Task, error) {

	var result []Task

	rows, err := db.Query("SELECT t.id, t.domain_id, t.name, t.weekly, t.description, t.count - ta.used AS available FROM tasks t LEFT JOIN (SELECT task_id, COUNT(*) AS used FROM task_assignments WHERE status != 'done_and_available' GROUP BY task_id) ta ON ta.task_id = t.id WHERE domain_id = $1", domain.ID)
	if err != nil {
		log.Printf("Error reading tasks: %q\n", err)
		return result, err
	}

	defer rows.Close()
	for rows.Next() {
		var t Task
		// results of math ops in postgres end up as int64 columns
		var taskCount int64

		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Weekly, &t.Description, &taskCount); err != nil {
			log.Printf("Error scanning task: %q", err)
			return result, err
		}
		if taskCount < 0 {
			log.Printf("Available count below 0 (%d) for domain %d\n", taskCount, domain.ID)
			return result, errors.New("DB state fail")
		}
		t.Count = uint32(taskCount)
		result = append(result, t)
	}

	return result, nil
}

func readTasksFromRows(rows *sql.Rows) ([]Task, error) {
	var result []Task

	defer rows.Close()
	for rows.Next() {
		var t Task

		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Weekly, &t.Description, &t.Count); err != nil {
			log.Printf("Error scanning task: %q", err)
			return result, err
		}
		result = append(result, t)
	}
	return result, nil
}

func GetTasksForDomain(domain Domain) ([]Task, error) {

	rows, err := db.Query("SELECT id, domain_id, name, weekly, description, count FROM tasks WHERE domain_id = $1", domain.ID)
	if err != nil {
		log.Printf("Error reading tasks for domain: %q", err)
		return nil, err
	}

	result, err := readTasksFromRows(rows)
	return result, err
}

func AssignmentInsert(assignment TaskAssignment) error {

	strDate := util.StrDateFromTime(assignment.AssignedDate.Time)
	_, err := db.Exec("INSERT INTO task_assignments (task_id, minion_id, assigned_on) VALUES($1,$2,$3)", assignment.Task.ID, assignment.MinionID, strDate)

	if err != nil {
		log.Printf("Error inserting new minion: %q", err)
		return err
	}
	return nil
}

func AssignmentUpdate(assignment TaskAssignment) error {

	strDateAssigned := util.StrDateFromTime(assignment.AssignedDate.Time)

	_, err := db.Exec("UPDATE task_assignments SET assigned_on = $1, status = $2 WHERE id = $3", strDateAssigned, assignment.Status, assignment.ID)

	if err != nil {
		log.Printf("Error updating assignment: %q", err)
		return err
	}
	return nil
}

func AssignmentDelete(assignment TaskAssignment) error {

	_, err := db.Exec("DELETE FROM task_assignments WHERE id = $1", assignment.ID)

	if err != nil {
		log.Printf("Error deleting assignment: %q", err)
		return err
	}
	return nil
}

func AssignmentRetrieve(taskAssignmentID int64) (TaskAssignment, error) {

	var result TaskAssignment

	row := db.QueryRow("SELECT id, task_id, assigned_on, status, CURRENT_DATE - assigned_on AS days_old FROM task_assignments WHERE id = $1", taskAssignmentID)

	if err := row.Scan(&result.ID, &result.Task.ID, &result.AssignedDate, &result.Status, &result.AgeInDays); err != nil {
		log.Printf("Error scanning assignment: %q", err)
		return result, err
	}

	return result, nil
}

// Retrieve all pending tasks for a minion, across all domains
func AssignmentRetrieveForMinion(minion Minion, includeCompleted bool) ([]TaskAssignment, error) {

	var result []TaskAssignment

	sql := "SELECT ta.id, task_id, assigned_on, CURRENT_DATE - assigned_on AS days_old, ta.status, t.domain_id, t.name, t.weekly, t.description FROM task_assignments AS ta LEFT JOIN tasks AS t ON ta.task_id = t.id WHERE status = 'pending' AND ta.minion_id = $1"
	if includeCompleted {
		sql = "SELECT ta.id, task_id, assigned_on, CURRENT_DATE - assigned_on AS days_old, ta.status, t.domain_id, t.name, t.weekly, t.description FROM task_assignments AS ta LEFT JOIN tasks AS t ON ta.task_id = t.id WHERE ta.minion_id = $1"
	}

	rows, err := db.Query(sql, minion.ID)
	if err != nil {
		log.Printf("Error reading pending tasks: %q", err)
		return result, err
	}

	defer rows.Close()
	for rows.Next() {
		var ta TaskAssignment

		if err := rows.Scan(&ta.ID, &ta.Task.ID, &ta.AssignedDate, &ta.AgeInDays, &ta.Status, &ta.Task.DomainID, &ta.Task.Name, &ta.Task.Weekly, &ta.Task.Description); err != nil {
			log.Printf("Error scanning task: %q", err)
			return result, err
		}
		result = append(result, ta)
	}

	return result, nil
}
