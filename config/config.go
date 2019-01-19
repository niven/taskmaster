package config

import (
	"log"
	"os"
)

var (
	environmentVarNames = []string{"TASKMASTER_OAUTH_CLIENT_SECRET", "PORT"}
	EnvironmentVars     = make(map[string]string)
)

func init() {

	for _, name := range environmentVarNames {
		value := os.Getenv(name)

		if value == "" {
			log.Fatalf("$%s must be set", name)
		}

		EnvironmentVars[name] = value
	}
}
