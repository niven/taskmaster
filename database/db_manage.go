package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	// have to import with an underbar alias since we need the init() to run
	_ "github.com/lib/pq"

	"github.com/niven/taskmaster/config"
)

var (
	db *sql.DB
)

const query_dir = "database"

func init() {

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)

	err := config.ReadEnvironmentVars()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// if we do db, err := foo() then this 'db' shadows the global one
	db, err = sql.Open("postgres", config.EnvironmentVars["DATABASE_URL"])

	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
}

func runQueries(filename string) {

	log.Printf("Running queries from: %s\n", filename)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Print(err)
		return
	}

	str := string(b)
	queries := strings.Split(str, "\n")

	for _, query := range queries {
		log.Printf("%s\n", query)

		// skip empty lines and SQL comments
		if strings.HasPrefix(query, "--") || len(query) == 0 {
			continue
		}

		result, err := db.Exec(query)
		if err != nil {
			log.Println(err)
			continue
		}

		count, err := result.RowsAffected()
		if err == nil {
			log.Printf("Rows affected: %d\n", count)
		} else {
			log.Println(err)
		}
	}

}

func main() {

	var update_point int

	row := db.QueryRow("SELECT EXISTS( SELECT 1 FROM information_schema.tables WHERE table_name = 'version')")

	var exists bool
	row.Scan(&exists)

	if !exists {
		update_point = -1
	} else {
		row = db.QueryRow("SELECT MAX(point) FROM version")
		row.Scan(&update_point)
	}

	log.Printf("Update point: %d\n", update_point)

	updates, _ := filepath.Glob(filepath.Join(query_dir, "update_*.sql"))
	if updates == nil {
		log.Println("Nothing to do.")
		return
	}

	re := regexp.MustCompile(`update_(\d+).sql`)
	todo := map[int]string{}
	var points []int

	for _, filename := range updates {
		match := re.FindStringSubmatch(filename)
		intval, _ := strconv.Atoi(match[1])
		todo[intval] = filename
		points = append(points, intval)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i] < points[j]
	})
	for _, point := range points {
		if point > update_point {
			runQueries(todo[point])
		}
	}
	log.Println("Done")
}
