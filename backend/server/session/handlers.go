package session

import "net/http"

// SessionHandler checks to see if the current request has a valid session
// Returns OK (200) if the session is valid, otherwise Unauthorized (401)
func SessionHandler(w http.ResponseWriter, req *http.Request) {
	if IsValidSession(w, req) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
