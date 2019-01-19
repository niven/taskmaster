package db

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/lib/pq" // have to import with an alias since we need the init() to run

	"github.com/niven/taskmaster/config"
	. "github.com/niven/taskmaster/data"
)

var (
	db *sql.DB
)

func init() {

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
		return errors.New("Could not insert minion")
	}

	id, err := result.LastInsertId()
	log.Printf("Created new minion with ID %d\n", id)
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

func ReadAllMinions() ([]Minion, error) {

	rows, err := db.Query("SELECT * FROM minions")
	if err != nil {
		log.Printf("Error reading minions: %q", err)
		return nil, errors.New("Could not select from table")
	}

	result := []Minion{}

	defer rows.Close()
	for rows.Next() {
		var m Minion

		if err := rows.Scan(&m.ID, &m.Email, &m.Name); err != nil {
			log.Printf("Error scanning minion: %q", err)
			return nil, errors.New("Could not scan from rows")
		}
		log.Printf("%+v\n", m)
		result = append(result, m)
	}

	return result, nil
}
