package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // have to import with an alias since we need the init() to run

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
		log.Printf("%+v\n", m)
		result = append(result, m)
	}

	return result, nil
}

func GetTasksForDomains(domains []Domain) ([]Task, []Task, Task, error) {

	var (
		late   []Task
		weekly []Task
		today  Task
	)

	return late, weekly, today, nil

}
