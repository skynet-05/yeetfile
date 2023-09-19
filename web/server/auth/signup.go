package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"yeetfile/db"
	"yeetfile/shared"
	"yeetfile/utils"
)

var MissingField = errors.New("missing username or email")

// Signup uses values from the Signup struct to complete registration of a new
// user. A hash is generated from the provided password and entered into the
// "users" db table.
func Signup(signup shared.Signup) (string, error) {
	// Email and password can both be empty (if the user only wants an
	// account number), but if one is provided, the other must be too.
	if utils.IsEitherEmpty(signup.Email, signup.Password) {
		return "", MissingField
	} else if len(signup.Email) == 0 {
		// User is only signing up for an account number
		return db.NewUser("", []byte(""))
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(signup.Password), 8)
		if err != nil {
			return "", err
		}

		return db.NewUser(signup.Email, hash)
	}
}
