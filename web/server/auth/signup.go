package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/mail"
)

var MissingField = errors.New("missing username or email")

// SignupWithEmail uses values from the Signup struct to complete registration
// of a new user. A hash is generated from the provided password and entered
// into the "users" db table.
func SignupWithEmail(signup shared.Signup) error {
	// Email and password cannot be empty
	if len(signup.Email) == 0 || len(signup.Password) == 0 {
		return MissingField
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(signup.Password), 8)
	if err != nil {
		return err
	}

	code, err := db.NewVerification(signup.Email, hash, false)
	if err != nil {
		return err
	}

	err = mail.SendVerificationEmail(code, signup.Email)
	return err
}

// SignupAccountIDOnly creates a new user with only an account ID as the user's
// login credential.
func SignupAccountIDOnly() (string, error) {
	return db.NewUser("", []byte(""))
}
