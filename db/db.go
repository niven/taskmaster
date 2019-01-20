package db

import (
	"database/sql"
	"log"

	// have to import with an underbar alias since we need the init() to run
	_ "github.com/lib/pq"

	"github.com/niven/taskmaster/config"
	. "github.com/niven/taskmaster/data"
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

	result, err := db.Exec("INSERT INTO minions (email,name) VALUES($1,$2)", email, name)

	if err != nil {
		log.Printf("Error inserting new minion: %q", err)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Created new minion with ID %d\n", id)
	} else {
		log.Printf("Error: %v\n", err)
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

func GetDomainsForMinion(m Minion) ([]Domain, error) {

	rows, err := db.Query("SELECT * FROM domains WHERE owner = $1", m.ID)
	if err != nil {
		return nil, err
	}

	var result []Domain
	defer rows.Close()
	for rows.Next() {
		var d Domain

		if err := rows.Scan(&d.ID, &d.Owner, &d.Name); err != nil {
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

func GetAllTasks(domain Domain) ([]Task, error) {
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
