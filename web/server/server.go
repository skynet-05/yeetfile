package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"yeetfile/b2"
	"yeetfile/crypto"
	"yeetfile/db"
	"yeetfile/shared"
	"yeetfile/utils"
)

var B2 b2.Auth

type router struct {
	routes map[string]http.HandlerFunc
}

type Metadata struct {
	Name       string `json:"name"`
	Chunks     int    `json:"chunks"`
	Password   string `json:"password"`
	Downloads  int    `json:"downloads"`
	Expiration string `json:"expiration"`
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

func home(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Yeetfile home page\n")
}

func uploadInit(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Error", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var meta Metadata
	err := decoder.Decode(&meta)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	key, salt, err := crypto.DeriveKey([]byte(meta.Password), nil)
	encodedKey := hex.EncodeToString(key[:])

	encName := crypto.EncryptChunk(key, []byte(meta.Name))
	b64Name := hex.EncodeToString(encName[:])

	id, _ := db.NewMetadata(meta.Chunks, b64Name, salt)
	b2Upload := db.InsertNewUpload(id)

	exp := utils.StrToDuration(meta.Expiration)
	db.SetFileExpiry(id, meta.Downloads, time.Now().Add(exp))

	if meta.Chunks == 1 {
		info, err := InitB2Upload()
		if err != nil {
			http.Error(w, "Unable to init file", http.StatusBadRequest)
			return
		}

		b2Upload.UpdateUploadValues(
			info.UploadURL,
			info.AuthorizationToken,
			info.BucketID)
	} else {
		info, err := InitLargeB2Upload(b64Name)
		if err != nil {
			http.Error(w, "Unable to init file", http.StatusBadRequest)
			return
		}

		b2Upload.UpdateUploadValues(
			info.UploadURL,
			info.AuthorizationToken,
			info.FileID)
	}

	// Return ID to user
	// TODO: Make this not weird
	_, _ = io.WriteString(w, fmt.Sprintf("%s|%s", id, encodedKey))
}

func uploadData(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Error", http.StatusMethodNotAllowed)
		return
	}

	chunkNum, _ := strconv.Atoi(req.Header.Get("Chunk"))
	key := crypto.KeyFromHex(req.Header.Get("Key"))

	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	upload, b2Values := PrepareUpload(id, key, chunkNum, data)
	done, err := upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	if done {
		path := utils.GenFilePath()
		if db.SetMetadataPath(id, path) {
			_, _ = io.WriteString(w, path)
		} else {
			http.Error(w, "Error generating file path", http.StatusInternalServerError)
		}
	}
}

func download(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	path := segments[len(segments)-1]

	decoder := json.NewDecoder(req.Body)
	var d DownloadRequest
	err := decoder.Decode(&d)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	metadata := db.RetrieveMetadataByPath(path)
	nameBytes, _ := hex.DecodeString(metadata.Name)
	key, _, _ := crypto.DeriveKey([]byte(d.Password), metadata.Salt)
	name, err := crypto.DecryptString(key, nameBytes)

	if err != nil {
		http.Error(w, "Incorrect password", http.StatusForbidden)
		return
	}

	response := shared.DownloadResponse{
		Name:   name,
		ID:     metadata.ID,
		Key:    hex.EncodeToString(key[:]),
		Chunks: metadata.Chunks,
	}

	jsonData, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

func downloadChunk(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")

	if len(segments) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := segments[len(segments)-2]
	chunk, _ := strconv.Atoi(segments[len(segments)-1])
	key := crypto.KeyFromHex(req.Header.Get("Key"))

	metadata := db.RetrieveMetadata(id)

	bytes := DownloadFile(metadata.B2ID, metadata.Length, chunk, key)
	_, _ = w.Write(bytes)
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
	r.routes["/d/*/*"] = downloadChunk

	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}

func init() {
	var err error
	B2, err = b2.AuthorizeAccount(
		os.Getenv("B2_BUCKET_KEY_ID"),
		os.Getenv("B2_BUCKET_KEY"))
	if err != nil {
		panic(err)
	}
}
