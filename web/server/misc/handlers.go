package misc

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"yeetfile/shared"
	"yeetfile/web/config"
	"yeetfile/web/static"
)

// UpHandler is used as the health check endpoint for load balancing, docker, etc.
func UpHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// WordlistHandler returns the set of words recommended by the EFF for generating
// secure passwords
func WordlistHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shared.EFFWordList); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// FileHandler uses the embedded files from staticFiles to return a file
// resource based on its name
func FileHandler(strip string, prepend string, files embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = prepend + req.URL.Path
		req.URL.Path = strings.Replace(req.URL.Path, fmt.Sprintf("/%s", config.VERSION), "", 1)
		path := strings.Split(req.URL.Path, "/")
		name := path[len(path)-1]

		w.Header().Set("Cache-Control", "max-age=86400")
		w.Header().Set("Expires", time.Now().Add(time.Hour*24).Format(http.TimeFormat))

		minFile, ok := static.MinifiedFiles[name]
		if ok {
			// Found minified file, return this instead
			if strings.HasSuffix(name, ".js") {
				w.Header().Set("Content-Type", "application/javascript")
			} else if strings.HasSuffix(name, ".css") {
				w.Header().Set("Content-Type", "text/css")
			}

			_, _ = w.Write(minFile)
			return
		}

		http.StripPrefix(
			strip,
			http.FileServer(http.FS(files)),
		).ServeHTTP(w, req)
	}
}
