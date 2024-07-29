package misc

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"yeetfile/shared"
	"yeetfile/shared/constants"
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
		req.URL.Path = strings.Replace(
			req.URL.Path,
			fmt.Sprintf("/%s", config.YeetFileConfig.Version), "", 1)
		path := strings.Split(req.URL.Path, "/")
		name := path[len(path)-1]

		w.Header().Set("Cache-Control", "private, max-age=604800")
		w.Header().Set("Expires", time.Now().Add(time.Hour*24*7).Format(http.TimeFormat))

		minFile, ok := static.MinifiedFiles[name]
		if ok {
			if strings.Contains(name, shared.DBFilename) {
				// New db.js requests should regenerate the
				// random pass
				defaultPassBytes := make([]byte, 32)
				if !config.IsDebugMode {
					_, _ = rand.Read(defaultPassBytes)
				}
				minFile = []byte(strings.ReplaceAll(
					string(minFile),
					constants.JSRandomSessionKey,
					hex.EncodeToString(defaultPassBytes)))
			}

			// Found minified file, return this instead
			if strings.HasSuffix(name, "js") {
				w.Header().Set("Content-Type", "application/javascript")
			} else if strings.HasSuffix(name, "css") {
				w.Header().Set("Content-Type", "text/css")
			}

			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")

			_, _ = w.Write(minFile)
			return
		}

		http.StripPrefix(
			strip,
			http.FileServer(http.FS(files)),
		).ServeHTTP(w, req)
	}
}
