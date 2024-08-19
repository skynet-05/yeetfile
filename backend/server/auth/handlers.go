package auth

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
	"yeetfile/backend/mail"
	"yeetfile/backend/server/html"
	"yeetfile/backend/server/session"
	"yeetfile/backend/server/transfer/vault"
	"yeetfile/backend/utils"
	"yeetfile/shared"
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

	userID, err := ValidateCredentials(login.Identifier, login.LoginKeyHash)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("User not found, or incorrect password"))
		return
	}

	protectedKey, publicKey, err := db.GetUserKeys(userID)
	if err != nil {
		http.Error(w, "Error retrieving user keys", http.StatusInternalServerError)
		return
	}

	_ = session.SetSession(userID, w, req)
	_ = json.NewEncoder(w).Encode(shared.LoginResponse{
		PublicKey:    publicKey,
		ProtectedKey: protectedKey,
	})
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

	// Check if server has a password
	if config.YeetFileConfig.PasswordHash != nil {
		err := bcrypt.CompareHashAndPassword(
			config.YeetFileConfig.PasswordHash,
			[]byte(signupData.ServerPassword))

		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("Missing or invalid server password"))
			return
		}
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
			utils.Logf("Error: %v\n", err)
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
			errMsg := "Error creating account"
			if err == db.UserAlreadyExists {
				errMsg = "User already exists"
			}
			status = http.StatusBadRequest
			response = shared.SignupResponse{
				Error: errMsg,
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

// AccountHandler handles fetching and returning the user's account information.
func AccountHandler(w http.ResponseWriter, req *http.Request, id string) {
	if !session.IsValidSession(req) {
		http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
		return
	}

	switch req.Method {
	case http.MethodDelete:
		var deleteAccount shared.DeleteAccount
		if json.NewDecoder(req.Body).Decode(&deleteAccount) != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("error decoding request"))
			return
		}

		accountID := deleteAccount.Identifier
		var err error
		if strings.Contains(deleteAccount.Identifier, "@") {
			accountID, err = db.GetUserIDByEmail(accountID)
		}

		if err != nil || accountID != id {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("error verifying account"))
			return
		}

		err = vault.DeleteVaultFolder(id, id, false)
		if err != nil {
			log.Printf("Error deleting user root folder: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.DeleteUser(id)
		if err != nil {
			log.Printf("Error deleting user: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = session.RemoveSession(w, req)
		w.WriteHeader(http.StatusOK)

		return
	case http.MethodGet:
		user, err := db.GetUserByID(id)
		if err != nil {
			log.Printf("Error fetching user by id: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(shared.AccountResponse{
			Email:            user.Email,
			StorageAvailable: user.StorageAvailable,
			StorageUsed:      user.StorageUsed,
			SendAvailable:    user.SendAvailable,
			SendUsed:         user.SendUsed,
			SubscriptionExp:  user.MemberExp,
		})
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
	accountValues, err := db.VerifyUser(email, code)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set(html.ErrorHeader, "Incorrect verification code")
		html.VerifyPageHandler(w, req, email)
		return
	}

	// Create new user
	id, err := db.NewUser(db.User{
		Email:        email,
		PasswordHash: accountValues.PasswordHash,
		ProtectedKey: accountValues.ProtectedKey,
		PublicKey:    accountValues.PublicKey,
	})

	if err != nil {
		w.Header().Set(html.ErrorHeader, "Error creating new account")
		html.VerifyPageHandler(w, req, email)
		return
	}

	err = db.NewRootFolder(id, accountValues.RootFolderKey)

	if err != nil {
		w.Header().Set(html.ErrorHeader, "Error initializing user vault")
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
	_, err := db.VerifyUser(verify.ID, verify.Code)
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
		PublicKey:    verify.PublicKey,
	})

	if err != nil {
		utils.Log(fmt.Sprintf("Bad request: %v\n", err))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Bad request"))
		return
	}

	err = db.NewRootFolder(verify.ID, verify.RootFolderKey)
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
			code, _ := db.NewVerification(
				shared.Signup{Identifier: forgot.Email},
				nil,
				true)
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
	_, err = db.VerifyUser(reset.Email, reset.Code)
	if utils.HandleError(w, err, http.StatusBadRequest, "unable to reset password") {
		return
	}

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

// PubKeyHandler handles requests for a YeetFile user's public key, which can
// be used for sharing files/folders with the user.
func PubKeyHandler(w http.ResponseWriter, req *http.Request, _ string) {
	userIdentifier := req.URL.Query().Get("user")

	var userID string
	var err error
	if len(userIdentifier) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	} else if strings.Contains(userIdentifier, "@") {
		userID, err = db.GetUserIDByEmail(userIdentifier)
	} else {
		userID = userIdentifier
		_, err = db.GetUserByID(userID)
	}

	if err != nil || len(userID) == 0 {
		utils.Logf("Error in user lookup for pub key: %v\n", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	pubKey, err := db.GetUserPubKey(userID)
	if err != nil {
		utils.Logf("Error fetching pub key: %v\n", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	jsonData, _ := json.Marshal(shared.PubKeyResponse{PublicKey: pubKey})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

// ProtectedKeyHandler returns the user's protected key (private key encrypted
// with their user key). This is used when updating the protected key when
// a user changes their email or password.
func ProtectedKeyHandler(w http.ResponseWriter, _ *http.Request, id string) {
	protectedKey, _, err := db.GetUserKeys(id)
	if err != nil {
		log.Printf("Error fetching user keys: %v\n", err)
		http.Error(w, "Error fetching protected key", http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(shared.ProtectedKeyResponse{
		ProtectedKey: protectedKey,
	})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonData)
}

// ChangePasswordHandler validates the user's old login information, and uses
// the ChangePassword request struct to update their login info and protected
// key with new values.
func ChangePasswordHandler(w http.ResponseWriter, req *http.Request, id string) {
	var changePassword shared.ChangePassword
	if json.NewDecoder(req.Body).Decode(&changePassword) != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Unable to decode request body"))
		return
	}

	userID, err := ValidateCredentials(id, changePassword.PrevLoginKeyHash)
	if err != nil || id != userID {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Incorrect password"))
		return
	}

	bcryptHash, err := bcrypt.GenerateFromPassword(
		changePassword.NewLoginKeyHash, 8)
	if err != nil {
		log.Printf("Error generating bcrypt hash: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.UpdateUserLogin(id, bcryptHash, changePassword.ProtectedKey)
	if err != nil {
		log.Printf("Error updating user login credentials: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// RecyclePaymentIDHandler handles replacing the user's current payment ID with
// a new value
func RecyclePaymentIDHandler(w http.ResponseWriter, req *http.Request, userID string) {
	user, err := db.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Error fetching user", http.StatusBadRequest)
		return
	}

	err = db.RotateUserPaymentID(user.PaymentID)
	if err != nil {
		http.Error(w, "Error recycling payment ID", http.StatusBadRequest)
		return
	}
}
