package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yeetfile/crypto"
	"yeetfile/db"
	"yeetfile/shared"
	"yeetfile/utils"
	"yeetfile/web/server/auth"
)

var (
	key   = []byte(utils.GenRandomString(16))
	store = sessions.NewCookieStore(key)
)

// home returns the homepage html if not logged in, otherwise the upload page
func home(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Yeetfile home page\n")
}

// signup uses data from the incoming POST request to create a new user. The
// data received must match the shared.Signup struct.
func signup(w http.ResponseWriter, req *http.Request) {
	var signup shared.Signup
	err := json.NewDecoder(req.Body).Decode(&signup)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = auth.Signup(signup)
	if err != nil {
		if err == db.UserAlreadyExists {
			w.WriteHeader(http.StatusConflict)
		} else if err == auth.MissingField {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func signupHTML(w http.ResponseWriter, req *http.Request) {
	// TODO: Signup html
}

// verify handles account verification using the link sent to a user's
// email immediately after signup.
func verify(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	token := req.URL.Query().Get("token")

	// Ensure the URL has the correct params for validation
	if len(email) == 0 || len(token) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if db.VerifyUser(email, token) {
		// TODO: Redirect to home/upload page?
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusForbidden)
}

// uploadInit handles a POST request to /u with the metadata required to set
// up a file for uploading. This is defined in the UploadMetadata struct.
func uploadInit(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var meta shared.UploadMetadata
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
	db.SetFileExpiry(id, meta.Downloads, time.Now().Add(exp).UTC())

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

// uploadData handles the process of uploading file chunks to the server, after
// having already initialized the file metadata beforehand.
func uploadData(w http.ResponseWriter, req *http.Request) {
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
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// download fetches metadata for downloading a file, such as the name of the
// file, the number of chunks, and the key for decrypting each chunk.
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

// downloadChunk downloads individual chunks of a file using the chunk num from
// the file path and the decryption key in the header.
// Ex: /d/abc123/2 -- download the second chunk of file with id "abc123"
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

	eof, bytes := DownloadFile(metadata.B2ID, metadata.Length, chunk, key)

	// If the file is finished downloading, decrease the download counter
	// for that file, and delete if 0 are remaining
	rem := -1
	if eof {
		exp := db.GetFileExpiry(metadata.ID)
		rem = db.DecrementDownloads(metadata.ID)

		if rem == 0 {
			db.DeleteFileByID(metadata.ID)
		}

		if rem >= 0 {
			w.Header().Set("Downloads", strconv.Itoa(rem))
		}
		w.Header().Set("Date", fmt.Sprintf("%s", exp.Date.String()))
	}

	_, _ = w.Write(bytes)
}

// Run defines maps URL paths to handlers for the server and begins listening
// on the configured port.
func Run(port string) {
	r := &router{
		routes: make(map[Route]http.HandlerFunc),
	}

	r.routes[Route{Path: "/", Method: http.MethodGet}] = home

	// Upload
	r.routes[Route{
		Path:   "/u",
		Method: http.MethodPost,
	}] = AuthMiddleware(uploadInit)
	r.routes[Route{
		Path:   "/u/*",
		Method: http.MethodPost,
	}] = AuthMiddleware(uploadData)

	// Download
	r.routes[Route{Path: "/d/*", Method: http.MethodPost}] = download
	r.routes[Route{Path: "/d/*/*", Method: http.MethodPost}] = downloadChunk

	// Account Management
	r.routes[Route{
		Path:   "/signup",
		Method: http.MethodPost,
	}] = LimiterMiddleware(signup)
	r.routes[Route{Path: "/signup", Method: http.MethodGet}] = signupHTML
	r.routes[Route{Path: "/verify", Method: http.MethodGet}] = verify
	//r.routes["/login"] = login
	//r.routes["/account"] = account

	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
