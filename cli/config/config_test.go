package config

import (
	"strings"
	"testing"
)

const session = "test_session"

func TestReadConfig(t *testing.T) {
	paths, err := setupTempConfigDir()
	if err != nil {
		t.Fatal("Failed to set up temporary config directories")
	}

	config, err := ReadConfig(paths)
	if err != nil {
		t.Fatal("Failed to read config")
	}

	if !strings.Contains(config.Server, "http") {
		t.Fatal("Invalid config server")
	}
}

func TestReadSession(t *testing.T) {
	paths, err := setupTempConfigDir()
	if err != nil {
		t.Fatal("Failed to set up temporary config directories")
	}

	err = paths.SetSession(session)
	if err != nil {
		t.Fatal("Failed to set user session")
	}

	readSession := paths.ReadSession()
	if len(readSession) == 0 {
		t.Fatal("Failed to read user session")
	} else if readSession != session {
		t.Fatalf("Unexpected session value\n"+
			"(expected %s, got %s)", session, readSession)
	}
}
