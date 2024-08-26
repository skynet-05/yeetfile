package utils

import (
	"bytes"
	"fmt"
	"testing"
)

func TestParseDownloadString(t *testing.T) {
	path := "abc123"
	secret := "secret.goes.here"
	downloadString := fmt.Sprintf("%s#%s", path, secret)

	parsedPath, parsedSecret, err := ParseDownloadString(downloadString)
	if err != nil {
		t.Fatalf("Error parsing download string: %v", err)
	}

	if len(parsedPath) == 0 || len(parsedSecret) == 0 {
		t.Fatalf("Error retrieving path and secret from download str")
	} else if parsedPath != path || !bytes.Equal([]byte(secret), parsedSecret) {
		t.Fatalf("Parsed path or secret values are incorrect")
	}

	invalidDownloadString := "invalid"
	_, _, err = ParseDownloadString(invalidDownloadString)
	if err == nil {
		t.Fatalf("Invalid download string was parsed without an error")
	}
}
