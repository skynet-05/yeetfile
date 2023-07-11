package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"yeetfile/db"
)

type router struct {
	routes map[string]http.HandlerFunc
}

type metadata struct {
	Name   string `json:"name"`
	Chunks int    `json:"chunks"`
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for path, handler := range r.routes {
		if matchPath(path, req.URL.Path) {
			handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
}

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

func home(w http.ResponseWriter, req *http.Request) {
	_, _ = io.WriteString(w, "Yeetfile home page\n")
}

func uploadInit(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Error", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var meta metadata
	err := decoder.Decode(&meta)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, _ := db.InsertMetadata(meta.Chunks, meta.Name)

	// Return ID to user
	_, _ = io.WriteString(w, id)
}

func uploadData(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		_, _ = io.WriteString(w, "Method not allowed.\n")
		return
	}

	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	// TODO: Process individual file chunks and ensure chunk num doesn't
	// exceed count stored in metadata
	metadata := db.RetrieveMetadata(id)
	_, _ = io.WriteString(w, "upload file data: "+metadata.ID)
}

func download(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	tag := segments[len(segments)-1]

	// TODO: Fetch file by tag and begin download
	_, _ = io.WriteString(w, "Yeetfile download: "+tag+"\n")
}

func Run(port string) {
	r := &router{
		routes: make(map[string]http.HandlerFunc),
	}

	r.routes["/"] = home
	r.routes["/home"] = home

	// Upload
	r.routes["/u"] = uploadInit
	r.routes["/u/*"] = uploadData

	// Download
	r.routes["/d/*"] = download

	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
