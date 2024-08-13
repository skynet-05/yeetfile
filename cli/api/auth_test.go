//go:build server_test

package api

import (
	"testing"
)

func TestValidSessions(t *testing.T) {
	_, err := UserA.context.GetSession()
	if err != nil {
		t.Fatalf("User A session error: %v\n", err)
	}

	_, err = UserB.context.GetSession()
	if err != nil {
		t.Fatalf("User B session error: %v\n", err)
	}
}
