package db

import (
	"database/sql"
	"log"
	"time"

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

	log.Printf("LM for '%s'\n", email)
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

	rows, err := db.Query("SELECT id, owner, name, last_reset_date FROM domains WHERE owner = $1", m.ID)
	if err != nil {
		return nil, err
	}

	var result []Domain
	defer rows.Close()
	for rows.Next() {
		var d Domain

		if err := rows.Scan(&d.ID, &d.Owner, &d.Name, &d.LastResetDate); err != nil {
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

func SaveTaskAssignment(taskID, minionID uint32, assignedOn time.Time) error {

	strDate := util.StrDateFromTime(assignedOn)
	_, err := db.Exec("INSERT INTO task_state (task_id, minion_id, assigned_on) VALUES($1,$2,$3)", taskID, minionID, strDate)

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

func GetTasksForDomain(domain Domain) ([]Task, error) {
	var result []Task

	rows, err := db.Query("SELECT t.id, t.domain_id, t.name, t.weekly, t.description, ts.assigned_on, ts.completed_on FROM tasks AS t LEFT JOIN task_state AS ts ON t.id = ts.task_id WHERE t.domain_id = $1", domain.ID)
	if err != nil {
		log.Printf("Error reading tasks: %q", err)
		return result, err
	}

	defer rows.Close()
	for rows.Next() {
		var t Task

		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Weekly, &t.Description, &t.AssignedDate, &t.CompletedDate); err != nil {
			log.Printf("Error scanning task: %q", err)
			return result, err
		}
		result = append(result, t)
	}

	return result, nil
}

// Retrieve all pending tasks for a minion, across all domains
func GetPendingTasksForMinion(minion Minion) ([]Task, error) {

	var tasks []Task

	rows, err := db.Query("SELECT ts.task_id, t.domain_id, t.name, t.weekly, t.description, ts.assigned_on FROM task_state AS ts LEFT JOIN tasks AS t ON ts.task_id = t.id WHERE completed_on IS NULL AND ts.minion_id = $1", minion.ID)
	if err != nil {
		log.Printf("Error reading pending tasks: %q", err)
		return tasks, err
	}

	defer rows.Close()
	for rows.Next() {
		var t Task

		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Weekly, &t.Description, &t.AssignedDate); err != nil {
			log.Printf("Error scanning task: %q", err)
			return tasks, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}
