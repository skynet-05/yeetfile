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
	"yeetfile/web/utils"
)

// LoginHandler handles a POST request to /login to log the user in.
func LoginHandler(w http.ResponseWriter, req *http.Request) {
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

	_ = SetSession(identifier, w, req)
	//http.Redirect(w, req, "/", http.StatusFound)
	w.WriteHeader(http.StatusOK)
}

// SignupHandler uses data from the incoming POST request to create a new user.
// The data received must match the shared.Signup struct.
func SignupHandler(w http.ResponseWriter, req *http.Request) {
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
		err = SetSession(id, w, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = io.WriteString(w, id)
	}
}

// VerifyHandler handles account verification using the link sent to a user's
// email immediately after signup.
func VerifyHandler(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	code := req.URL.Query().Get("code")

	// Ensure the request has the correct params for verification, otherwise
	// it should return the HTML for the verification page
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

	_ = SetSession(id, w, req)
	//http.Redirect(w, req, "/", http.StatusFound)
	w.WriteHeader(http.StatusOK)
}

// SessionHandler checks to see if the current request has a valid session
// Returns OK (200) if the session is valid, otherwise Unauthorized (401)
func SessionHandler(w http.ResponseWriter, req *http.Request) {
	if IsValidSession(req) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// LogoutHandler handles a PUT request to /logout to log the user out of their
// current session.
func LogoutHandler(w http.ResponseWriter, req *http.Request) {
	err := RemoveSession(w, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
