package auth

import (
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"strings"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/server/html"
	"yeetfile/web/server/session"
	"yeetfile/web/utils"
)

// LoginHandler handles a POST request to /login to log the user in.
func LoginHandler(w http.ResponseWriter, req *http.Request) {
	var identifier string
	var password []byte

	_ = req.ParseForm()

	if req.FormValue("email") != "" {
		identifier = req.FormValue("email")
		password = []byte(req.FormValue("password"))
	} else {
		var loginFields shared.Login
		err := json.NewDecoder(req.Body).Decode(&loginFields)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		identifier = loginFields.Identifier
		password = []byte(loginFields.Password)
	}

	if strings.Contains(identifier, "@") {
		pwHash, err := db.GetUserPasswordHashByEmail(identifier)
		if err != nil || bcrypt.CompareHashAndPassword(pwHash, password) != nil {
			w.Header().Set(html.ErrorHeader, "User not found, or incorrect password")
			html.LoginPageHandler(w, req)
			return
		}

		identifier, err = db.GetUserIDByEmail(identifier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		pwHash, err := db.GetUserPasswordHashByID(identifier)
		if (pwHash != nil && len(pwHash) != 0) || err != nil || !db.UserIDExists(identifier) {
			w.Header().Set(html.ErrorHeader, "Account not found")
			html.LoginPageHandler(w, req)
			return
		}
	}

	_ = session.SetSession(identifier, w, req)
	req.Method = http.MethodGet
	http.Redirect(w, req, "/", http.StatusMovedPermanently)
}

// SignupHandler uses data from the incoming POST request to create a new user.
// The data received must match the shared.Signup struct.
func SignupHandler(w http.ResponseWriter, req *http.Request) {
	var signupData shared.Signup
	if json.NewDecoder(req.Body).Decode(&signupData) != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Unable to parse request"))
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
		id, err = SignupAccountIDOnly()
	} else {
		// Need email verification before finishing with signup
		err = SignupWithEmail(signupData)
	}

	if err != nil {
		if errors.Is(err, db.UserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("User already exists"))
		} else if errors.Is(err, MissingField) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad request"))
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Server error"))
		}
		return
	} else if len(signupData.Email) == 0 {
		err = session.SetSession(id, w, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = io.WriteString(w, id)
	}
}

// AccountHandler handles fetching the user's data and displaying a web page for
// managing their account (web only)
func AccountHandler(w http.ResponseWriter, req *http.Request) {
	if !session.IsValidSession(req) {
		http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
	} else {
		s, _ := session.GetSession(req)
		id := session.GetSessionUserID(s)
		user, err := db.GetUserByID(id)
		if err != nil {
			log.Printf("Error fetching user by id: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Method == http.MethodGet {
			html.AccountPageHandler(w, req, user)
			return
		}
	}
}

// VerifyHandler handles account verification using the link sent to a user's
// email immediately after signup.
func VerifyHandler(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	code := req.URL.Query().Get("code")

	// Ensure the request has the correct params for verification, otherwise
	// it should return the HTML verification page
	if len(email) == 0 || len(code) == 0 {
		if len(email) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		html.VerifyPageHandler(w, req, email)
		return
	}

	// Verify user verification code and fetch password hash
	pwHash, err := db.VerifyUser(email, code)
	if err != nil {
		w.Header().Set(html.ErrorHeader, "Incorrect verification code")
		html.VerifyPageHandler(w, req, email)
		return
	}

	// Create new user
	id, err := db.NewUser(email, pwHash)
	if err != nil {
		w.Header().Set(html.ErrorHeader, "Server error")
		html.VerifyPageHandler(w, req, email)
		return
	}

	// Remove verification entry
	_ = db.DeleteVerification(email)

	_ = session.SetSession(id, w, req)
	http.Redirect(w, req, "/account", http.StatusMovedPermanently)
}

// LogoutHandler handles a PUT request to /logout to log the user out of their
// current session.
func LogoutHandler(w http.ResponseWriter, req *http.Request) {
	err := session.RemoveSession(w, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
}
