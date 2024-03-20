package auth

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/mail"
	"yeetfile/web/server/html"
	"yeetfile/web/server/session"
	"yeetfile/web/utils"
)

// LoginHandler handles a POST request to /login to log the user in.
func LoginHandler(w http.ResponseWriter, req *http.Request) {
	var err error

	var login shared.Login
	login, err = utils.GetStructFromFormOrJSON(&login, req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	identifier := login.Identifier
	keyHash := login.LoginKeyHash

	if strings.Contains(login.Identifier, "@") {
		pwHash, err := db.GetUserPasswordHashByEmail(identifier)
		if err != nil || bcrypt.CompareHashAndPassword(pwHash, keyHash) != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("User not found, or incorrect password"))
			return
		}

		identifier, err = db.GetUserIDByEmail(identifier)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
			return
		}
	} else {
		pwHash, err := db.GetUserPasswordHashByID(identifier)
		pwError := bcrypt.CompareHashAndPassword(pwHash, keyHash)
		if err != nil || !db.UserIDExists(identifier) || pwError != nil {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("User not found, or incorrect password"))
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

	var response shared.SignupResponse
	status := http.StatusOK

	if len(signupData.Identifier) == 0 {
		// No email, so this is an account ID only signup
		id, captcha, err := SignupAccountIDOnly()
		if err != nil {
			status = http.StatusBadRequest
			response = shared.SignupResponse{
				Error: "Error creating account ID",
			}
		} else {
			response = shared.SignupResponse{
				Identifier: id,
				Captcha:    captcha,
			}
		}
	} else {
		// Email signup
		err := SignupWithEmail(signupData)
		if err != nil {
			status = http.StatusBadRequest
			response = shared.SignupResponse{
				Error: "Error creating account ID",
			}
		} else {
			response = shared.SignupResponse{
				Identifier: signupData.Identifier,
			}
		}
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// AccountHandler handles fetching the user's data and displaying a web page for
// managing their account (web only)
func AccountHandler(w http.ResponseWriter, req *http.Request) {
	if !session.IsValidSession(req) {
		http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
		return
	}

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

// VerifyEmailHandler handles account verification using the link sent to a user's
// email immediately after signup.
func VerifyEmailHandler(w http.ResponseWriter, req *http.Request) {
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
	pwHash, protectedKey, err := db.VerifyUser(email, code)
	if err != nil {
		w.Header().Set(html.ErrorHeader, "Incorrect verification code")
		html.VerifyPageHandler(w, req, email)
		return
	}

	// Create new user
	id, err := db.NewUser(db.User{
		Email:        email,
		PasswordHash: pwHash,
		ProtectedKey: protectedKey,
	})

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

// VerifyAccountHandler handles account verification using the CAPTCHA displayed
// to the user containing a multi-digit code.
func VerifyAccountHandler(w http.ResponseWriter, req *http.Request) {
	var verify shared.VerifyAccount
	if json.NewDecoder(req.Body).Decode(&verify) != nil {
		utils.Log("Unable to parse request")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Unable to parse request"))
		return
	} else if utils.IsStructMissingAnyField(verify) {
		utils.Log("Missing required fields for verification")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Unable to parse request"))
		return
	}

	// Verify user verification code
	_, _, err := db.VerifyUser(verify.ID, verify.Code)
	if err != nil {
		utils.Log("Incorrect verification code")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Incorrect verification code"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword(verify.LoginKeyHash, 8)

	_, err = db.NewUser(db.User{
		ID:           verify.ID,
		ProtectedKey: verify.ProtectedKey,
		PasswordHash: hash,
	})

	if err != nil {
		utils.Log(fmt.Sprintf("Bad request: %v\n", err))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
		return
	}

	// Remove verification entry
	_ = db.DeleteVerification(verify.ID)
	_ = session.SetSession(verify.ID, w, req)
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

// ForgotPasswordHandler handles a GET request for returning a form for the user
// to fill out to recover their password, or a POST request for submitting the
// request to reset their password.
func ForgotPasswordHandler(w http.ResponseWriter, req *http.Request) {
	if session.IsValidSession(req) {
		http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
		return
	}

	if req.Method == http.MethodGet {
		html.ForgotPageHandler(w, req, "")
		return
	} else if req.Method == http.MethodPost {
		_ = req.ParseForm()

		var forgot shared.ForgotPassword
		forgot, err := utils.GetStructFromFormOrJSON(&forgot, req)

		id, err := db.GetUserIDByEmail(forgot.Email)
		if err == nil && len(id) > 0 && len(forgot.Email) > 0 {
			code, _ := db.NewVerification(forgot.Email, nil, nil, true)
			_ = mail.SendResetEmail(code, forgot.Email)
		}

		redirect := fmt.Sprintf("/forgot?email=%s", forgot.Email)
		http.Redirect(w, req, redirect, http.StatusSeeOther)
	}
}

// ResetPasswordHandler receives a request with a verification code, email,
// and new password to reset a user's password.
func ResetPasswordHandler(w http.ResponseWriter, req *http.Request) {
	var reset shared.ResetPassword
	reset, err := utils.GetStructFromFormOrJSON(&reset, req)
	if err != nil {
		fmt.Printf("%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errorMsg := ""
	_, _, err = db.VerifyUser(reset.Email, reset.Code)
	if err != nil {
		errorMsg = "Incorrect verification code"
	} else if reset.Password != reset.ConfirmPassword {
		errorMsg = "Passwords don't match"
	}

	if len(errorMsg) > 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set(html.ErrorHeader, errorMsg)
		html.ForgotPageHandler(w, req, reset.Email)
		return
	}

	_ = db.DeleteVerification(reset.Email)
	hash, _ := bcrypt.GenerateFromPassword([]byte(reset.Password), 8)
	_ = db.SetNewPassword(reset.Email, hash)

	w.Header().Set(html.SuccessHeader, "Password successfully reset!")
	html.LoginPageHandler(w, req)
}
