package misc

import (
	"embed"
	"encoding/json"
	"net/http"
	"yeetfile/web/utils"
)

// UpHandler is used as the health check endpoint for load balancing, docker, etc.
func UpHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// WordlistHandler returns the set of words recommended by the EFF for generating
// secure passwords
func WordlistHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(utils.EFFWordList); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// FileHandler uses the embedded files from staticFiles to return a file
// resource based on its name
func FileHandler(staticFiles embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, req)
	}
}
