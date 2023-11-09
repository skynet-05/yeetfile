package server

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/server/auth"
	"yeetfile/web/server/payments"
	"yeetfile/web/templates"
	"yeetfile/web/utils"
)

const (
	POST   = http.MethodPost
	GET    = http.MethodGet
	PUT    = http.MethodPut
	DELETE = http.MethodDelete
)

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
	var signupData shared.Signup
	if json.NewDecoder(req.Body).Decode(&signupData) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var id string
	var err error

	if utils.IsEitherEmpty(signupData.Email, signupData.Password) {
		// If email is empty but not the password (or vice versa) the
		// request is invalid.
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
		return
	} else if len(signupData.Email) == 0 {
		// No email (or password), so this is an account ID only signup
		id, err = auth.SignupAccountIDOnly()
	} else {
		// Need email verification before finishing with signup
		err = auth.SignupWithEmail(signupData)
	}

	if err != nil {
		if errors.Is(err, db.UserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("User already exists"))
		} else if errors.Is(err, auth.MissingField) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad request"))
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Server error"))
		}
		return
	} else if len(signupData.Email) == 0 {
		err = auth.SetSession(id, w, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = io.WriteString(w, id)
	}
}

func signupHTML(w http.ResponseWriter, req *http.Request) {
	// TODO: Signup html
}

// verify handles account verification using the link sent to a user's
// email immediately after signup.
func verify(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	code := req.URL.Query().Get("code")

	// Ensure the URL has the correct params for validation
	if len(email) == 0 || len(code) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Verify user verification code and fetch password hash
	pwHash, err := db.VerifyUser(email, code)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Create new user
	id, err := db.NewUser(email, pwHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Remove verification entry
	_ = db.DeleteVerification(email)

	_ = auth.SetSession(id, w, req)
	//http.Redirect(w, req, "/", http.StatusFound)
	w.WriteHeader(http.StatusOK)
}

// login handles a POST request to /login to log the user in.
func login(w http.ResponseWriter, req *http.Request) {
	var loginFields shared.Login
	err := json.NewDecoder(req.Body).Decode(&loginFields)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	identifier := loginFields.Identifier
	password := []byte(loginFields.Password)

	if strings.Contains(loginFields.Identifier, "@") {
		pwHash, err := db.GetUserPasswordHashByEmail(identifier)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if bcrypt.CompareHashAndPassword(pwHash, password) != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		identifier, err = db.GetUserIDByEmail(identifier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		if !db.UserIDExists(identifier) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	_ = auth.SetSession(identifier, w, req)
	//http.Redirect(w, req, "/", http.StatusFound)
	w.WriteHeader(http.StatusOK)
}

// checkSession checks to see if the current request has a valid session (return
// 200) or not (401)
func checkSession(w http.ResponseWriter, req *http.Request) {
	if auth.IsValidSession(req) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// logout handles a PUT request to /logout to log the user out of their
// current session.
func logout(w http.ResponseWriter, req *http.Request) {
	err := auth.RemoveSession(w, req)
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
func fileHandler(staticFiles embed.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		http.FileServer(http.FS(staticFiles)).ServeHTTP(w, req)
	}
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
	r := &router{
		routes: make(map[Route]http.HandlerFunc),
	}

	r.routes[Route{Path: "/", Method: GET}] = home
	r.routes[Route{Path: "/upload", Method: GET}] = home

	// Upload
	r.routes[Route{Path: "/u", Method: POST}] = AuthMiddleware(uploadInit)
	r.routes[Route{Path: "/u/*/*", Method: POST}] = AuthMiddleware(uploadData)

	// Download
	r.routes[Route{Path: "/*", Method: GET}] = downloadHTML
	r.routes[Route{Path: "/d/*", Method: GET}] = download
	r.routes[Route{Path: "/d/*/*", Method: GET}] = downloadChunk

	// Account Management
	r.routes[Route{Path: "/signup", Method: POST}] = LimiterMiddleware(signup)
	r.routes[Route{Path: "/signup", Method: GET}] = signupHTML
	r.routes[Route{Path: "/verify", Method: GET}] = verify
	r.routes[Route{Path: "/login", Method: POST}] = login
	r.routes[Route{Path: "/logout", Method: PUT}] = logout
	//r.routes["/account"] = account

	// Misc
	r.routes[Route{Path: "/static/*/*", Method: GET}] = fileHandler(files)
	r.routes[Route{Path: "/wordlist", Method: GET}] = wordlist
	r.routes[Route{Path: "/faq", Method: GET}] = faqHTML
	r.routes[Route{Path: "/up", Method: GET}] = up
	r.routes[Route{Path: "/session", Method: GET}] = checkSession

	// Payments
	r.routes[Route{Path: "/stripe", Method: POST}] = payments.StripeWebhook

	// Reserve endpoints to protect against bad wildcard matches
	for route := range r.routes {
		endpoint := strings.Split(route.Path, "/")[1]
		if len(endpoint) > 0 && endpoint != "*" {
			r.reserved = append(r.reserved, endpoint)
		}
	}

	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
