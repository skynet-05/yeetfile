package server

import (
	"net/http"
	"strings"
)

type router struct {
	routes map[string]http.HandlerFunc
}

// ServeHTTP finds the proper routing handler for the provided path.
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for path, handler := range r.routes {
		if matchPath(path, req.URL.Path) {
			handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
}

// matchPath takes a URL path and determines if it's a match for a particular
// handler. This allows differentiating between two paths where the only
// difference is a wildcard (i.e. "/u" and "/u/*" for uploadInit and uploadData)
func matchPath(pattern, path string) bool {
	parts := strings.Split(pattern, "/")
	segments := strings.Split(path, "/")

	if len(parts) != len(segments) {
		return false
	}

	for i, part := range parts {
		if part == "*" {
			continue
		}

		if part != segments[i] {
			return false
		}
	}

	return true
}
