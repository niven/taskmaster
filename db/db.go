package db

import (
	"database/sql"
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

func SaveTaskAssignment(assignment TaskAssignment) error {

	strDate := util.StrDateFromTime(assignment.AssignedDate.Time)
	_, err := db.Exec("INSERT INTO task_state (task_id, minion_id, assigned_on) VALUES($1,$2,$3)", assignment.Task.ID, assignment.MinionID, strDate)

	if err != nil {
		log.Printf("Error inserting new minion: %q", err)
		return err
	}
	return nil
}

func ResetAllCompletedTasks(domain Domain) error {

	result, err := db.Exec("DELETE FROM task_state WHERE completed_on IS NOT NULL AND domain_id = $1", domain.ID)
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

	rows, err := db.Query("SELECT id, domain_id, name, weekly, description, count - COUNT(ts.task_id) FROM tasks t LEFT JOIN task_state ts ON t.id=ts.task_id WHERE domain_id = $1 GROUP BY id;", domain.ID)
	if err != nil {
		log.Printf("Error reading tasks: %q", err)
		return result, err
	}

	defer rows.Close()
	for rows.Next() {
		var t Task
		var taskCount int64

		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Weekly, &t.Description, &taskCount); err != nil {
			log.Printf("Error scanning task: %q", err)
			return result, err
		}
		t.Count = uint32(taskCount)
		result = append(result, t)
	}

	return result, nil
}

func GetTasksForDomain(domain Domain) ([]Task, error) {

	var result []Task

	rows, err := db.Query("SELECT id, domain_id, name, weekly, description, count FROM tasks WHERE domain_id = $1", domain.ID)
	if err != nil {
		log.Printf("Error reading tasks: %q", err)
		return result, err
	}

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

// Retrieve all pending tasks for a minion, across all domains
func GetPendingTasksForMinion(minion Minion) ([]TaskAssignment, error) {

	var result []TaskAssignment

	rows, err := db.Query("SELECT task_id, assigned_on FROM task_state AS ts LEFT JOIN tasks AS t ON ts.task_id = t.id WHERE completed_on IS NULL AND ts.minion_id = $1", minion.ID)
	if err != nil {
		log.Printf("Error reading pending tasks: %q", err)
		return result, err
	}

	defer rows.Close()
	taskIDs := make(map[uint32]bool)
	for rows.Next() {
		var ta TaskAssignment
		var id uint32

		if err := rows.Scan(&id, &ta.AssignedDate); err != nil {
			log.Printf("Error scanning task: %q", err)
			return result, err
		}
		taskIDs[id] = true
		result = append(result, ta)
	}

	return result, nil
}
