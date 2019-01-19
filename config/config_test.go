package config

import (
	"os"
	"testing"
)

func TestReadSetEnvironment(t *testing.T) {
	for _, name := range environmentVarNames {
		os.Setenv(name, "TEST")
	}
	err := ReadEnvironmentVars()
	if err != nil {
		t.Fail()
	}
}

func TestReadUnsetEnvironment(t *testing.T) {
	for _, name := range environmentVarNames {
		os.Unsetenv(name)
	}
	err := ReadEnvironmentVars()
	if err == nil {
		t.Fail()
	}

}
