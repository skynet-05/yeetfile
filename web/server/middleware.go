package server

import "net/http"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		session, _ := store.Get(req, "session")
		if ok, found := session.Values["auth"].(bool); !ok || !found {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// If the user is authenticated, call the next handler
		next.ServeHTTP(w, req)
	}

	return handler
}
