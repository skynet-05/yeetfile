package utils

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestGeneratePassphrase(t *testing.T) {
	passphrase := GeneratePassphrase()
	words := strings.Split(passphrase, separator)
	if len(words) != 3 {
		t.Fatalf("Unexpected passphrase length\n"+
			"(expected: %d, got: %d)", 3, len(words))
	}

	// Check passphrase length against (arbitrary) size
	if len(passphrase) < 10 {
		t.Fatalf("Passphrase length is too short")
	}
}

func TestParseDownloadString(t *testing.T) {
	path := "abc123"
	pepper := "pepper.goes.here"
	downloadString := fmt.Sprintf("%s#%s", path, pepper)

	parsedPath, parsedPepper, err := ParseDownloadString(downloadString)
	if err != nil {
		t.Fatalf("Error parsing download string: %v", err)
	}

	if len(parsedPath) == 0 || len(parsedPepper) == 0 {
		t.Fatalf("Error retrieving path and pepper from download str")
	} else if parsedPath != path || !bytes.Equal([]byte(pepper), parsedPepper) {
		t.Fatalf("Parsed path or pepper values are incorrect")
	}

	invalidDownloadString := "invalid"
	_, _, err = ParseDownloadString(invalidDownloadString)
	if err == nil {
		t.Fatalf("Invalid download string was parsed without an error")
	}
}
