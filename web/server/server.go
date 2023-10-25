package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yeetfile/db"
	"yeetfile/payments"
	"yeetfile/shared"
	"yeetfile/utils"
	"yeetfile/web/server/auth"
	"yeetfile/web/templates"
)

var staticFiles embed.FS

// home returns the homepage html if not logged in, otherwise the upload page
func home(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.UploadHTML,
		templates.Template{LoggedIn: true},
	)
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

	id, err := auth.Signup(signup)
	if err != nil {
		if errors.Is(err, db.UserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("User already exists"))
		} else if errors.Is(err, auth.MissingField) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad request"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Server error"))
		}
		return
	}

	session, _ := GetSession(req)
	session.Values["authenticated"] = true
	err = session.Save(req, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = io.WriteString(w, id)
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

	//if db.VerifyUser(email, token) {
	//	// TODO: Redirect to home/upload page?
	//	w.WriteHeader(http.StatusOK)
	//	return
	//}

	w.WriteHeader(http.StatusForbidden)
}

// logout handles a PUT request to /logout to log the user out of their
// current session.
func logout(w http.ResponseWriter, req *http.Request) {
	session, _ := GetSession(req)

	session.Values["authenticated"] = false
	err := session.Save(req, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// uploadInit handles a POST request to /u with the metadata required to set
// up a file for uploading. This is defined in the UploadMetadata struct.
func uploadInit(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var meta shared.UploadMetadata
	err := decoder.Decode(&meta)
	if err != nil {
		log.Printf("%v\n", req.Body)
		log.Printf("Error: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, _ := db.NewMetadata(meta.Chunks, meta.Name, meta.Salt)
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
		info, err := InitLargeB2Upload(meta.Name)
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
	_, _ = io.WriteString(w, id)
}

// uploadData handles the process of uploading file chunks to the server, after
// having already initialized the file metadata beforehand.
func uploadData(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-2]
	chunkNum, err := strconv.Atoi(segments[len(segments)-1])
	if err != nil {
		http.Error(w, "Invalid upload URL", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	upload, b2Values := PrepareUpload(id, chunkNum, data)
	done, err := upload.Upload(b2Values)

	if err != nil {
		http.Error(w, "Upload error", http.StatusBadRequest)
		return
	}

	if done {
		_, _ = io.WriteString(w, id)
	}
}

// downloadHTML returns the HTML page for downloading a file
func downloadHTML(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.DownloadHTML,
		templates.Template{LoggedIn: true},
	)
}

// download fetches metadata for downloading a file, such as the name of the
// file, the number of chunks, and the key for decrypting each chunk.
func download(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(req.URL.Path, "/")
	id := segments[len(segments)-1]

	metadata := db.RetrieveMetadata(id)
	expiry := db.GetFileExpiry(id)

	response := shared.DownloadResponse{
		Name:       metadata.Name,
		ID:         metadata.ID,
		Chunks:     metadata.Chunks,
		Salt:       metadata.Salt,
		Size:       metadata.Length,
		Downloads:  expiry.Downloads,
		Expiration: expiry.Date,
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

	metadata := db.RetrieveMetadata(id)

	eof, bytes := DownloadFile(metadata.B2ID, metadata.Length, chunk)

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

// fileHandler uses the embedded files from staticFiles to return a file
// resource based on its name
func fileHandler(w http.ResponseWriter, req *http.Request) {
	http.FileServer(http.FS(staticFiles)).ServeHTTP(w, req)
}

// wordlist returns the set of words recommended by the EFF for generating
// secure passwords
func wordlist(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(utils.EFFWordList); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

// faqHTML returns the FAQ HTML page
func faqHTML(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.FaqHTML,
		templates.Template{LoggedIn: true},
	)
}

// up is used as the health check endpoint for load balancing, docker, etc.
func up(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Run defines maps URL paths to handlers for the server and begins listening
// on the configured port.
func Run(port string, files embed.FS) {
	staticFiles = files

	r := &router{
		routes: make(map[Route]http.HandlerFunc),
	}

	r.routes[Route{Path: "/", Method: http.MethodGet}] = home
	r.routes[Route{Path: "/upload", Method: http.MethodGet}] = home

	// Upload
	r.routes[Route{
		Path:   "/u",
		Method: http.MethodPost,
	}] = AuthMiddleware(uploadInit)
	r.routes[Route{
		Path:   "/u/*/*",
		Method: http.MethodPost,
	}] = AuthMiddleware(uploadData)

	// Download
	r.routes[Route{Path: "/*", Method: http.MethodGet}] = downloadHTML
	r.routes[Route{Path: "/d/*", Method: http.MethodGet}] = download
	r.routes[Route{Path: "/d/*/*", Method: http.MethodGet}] = downloadChunk

	// Account Management
	r.routes[Route{
		Path:   "/signup",
		Method: http.MethodPost,
	}] = LimiterMiddleware(signup)
	r.routes[Route{Path: "/signup", Method: http.MethodGet}] = signupHTML
	r.routes[Route{Path: "/verify", Method: http.MethodGet}] = verify
	r.routes[Route{Path: "/logout", Method: http.MethodPut}] = logout
	//r.routes["/login"] = login
	//r.routes["/account"] = account

	// Misc
	r.routes[Route{Path: "/static/*/*", Method: http.MethodGet}] = fileHandler
	r.routes[Route{Path: "/wordlist", Method: http.MethodGet}] = wordlist
	r.routes[Route{Path: "/faq", Method: http.MethodGet}] = faqHTML
	r.routes[Route{Path: "/up", Method: http.MethodGet}] = up

	// Payments
	r.routes[Route{Path: "/stripe", Method: http.MethodPost}] = payments.StripeWebhook

	// Reserve endpoints to protect against bad wildcard matches
	for route := range r.routes {
		endpoint := strings.Split(route.Path, "/")[1]
		if len(endpoint) > 0 && endpoint != "*" {
			reservedEndpoints = append(reservedEndpoints, endpoint)
		}
	}

	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
