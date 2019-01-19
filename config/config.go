package config

import (
	"fmt"
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

func ReadEnvironmentVars() error {

	for _, name := range environmentVarNames {

		value := os.Getenv(name)

		if value == "" {
			return fmt.Errorf("$%s must be set", name)
		}
		log.Printf("$%s='%s'\n", name, value)
		EnvironmentVars[name] = value
	}

	return nil
}
