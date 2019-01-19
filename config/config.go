package config

import (
	"log"
	"os"
)

var (
	environmentVarNames = []string{
		"DATABASE_URL",
		"PORT",
		"TASKMASTER_OAUTH_CLIENT_SECRET",
	}
	EnvironmentVars = make(map[string]string)
)

func init() {

	for _, name := range environmentVarNames {
		log.Printf("Reading %s\n", name)
		value := os.Getenv(name)

		if value == "" {
			log.Fatalf("$%s must be set", name)
		}
		EnvironmentVars[name] = value
	}
}
