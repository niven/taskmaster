package config

import (
	"fmt"
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
		EnvironmentVars[name] = value
	}

	return nil
}
