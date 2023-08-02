package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"yeetfile/db"
	"yeetfile/shared"
)

var MissingField = errors.New("missing username or email")

// Signup uses values from the Signup struct to complete registration of a new
// user. A hash is generated from the provided password and entered into the
// "users" db table.
func Signup(signup shared.Signup) error {
	if len(signup.Email) == 0 && len(signup.Username) == 0 {
		return MissingField
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(signup.Password), 8)
	if err != nil {
		return err
	}

	// TODO: Send email verification?
	err = db.NewUser(signup.Email, signup.Username, hash)
	if err != nil {
		return err
	}

	return nil
}
