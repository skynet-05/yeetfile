package auth

import (
	"golang.org/x/crypto/bcrypt"
	"yeetfile/db"
)

type Signup struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup uses values from the Signup struct to complete registration of a new
// user. A hash is generated from the provided password and entered into the
// "users" db table.
func (signup Signup) Signup() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(signup.Password), 8)
	if err != nil {
		return err
	}

	// TODO: Send email verification
	err = db.NewUser(signup.Email, hash)
	if err != nil {
		return err
	}

	return nil
}
