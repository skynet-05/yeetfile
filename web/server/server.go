package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"yeetfile/crypto"
	"yeetfile/db"
)

type router struct {
	routes map[string]http.HandlerFunc
}

type metadata struct {
	Name     string `json:"name"`
	Chunks   int    `json:"chunks"`
	Password string `json:"password"`
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

	key, salt, err := crypto.DeriveKey([]byte(meta.Password), nil)
	encodedKey := base64.StdEncoding.EncodeToString(key[:])

	id, _ := db.InsertMetadata(meta.Chunks, meta.Name, salt)

	// Return ID to user
	// TODO: Make this not weird
	_, _ = io.WriteString(w, fmt.Sprintf("%s|%s", id, encodedKey))
}

func uploadData(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Error", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Use chunk num + total chunks to determine which upload method
	// has to be used
	//chunkNum := req.Header.Get("Chunk")

	key := req.Header.Get("Key")
	decodedKey, _ := base64.StdEncoding.DecodeString(key)
	var keyBytes [crypto.KEY_SIZE]byte
	copy(keyBytes[:], decodedKey[:crypto.KEY_SIZE])

	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	// TODO: Process individual file chunks and ensure chunk num doesn't
	// exceed count stored in metadata
	metadata := db.RetrieveMetadata(id)
	upload := FileUpload{
		data:     data,
		filename: metadata.Name,
		key:      keyBytes,
		salt:     metadata.Salt,
	}
	upload.UploadFile(0)
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
