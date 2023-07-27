package db

import (
	"errors"
	"log"
	"yeetfile/utils"
)

var UserAlreadyExists = errors.New("user already exists")

// NewUser creates a new user in the "users" table, ensuring that the email
// provided is not already in use.
func NewUser(email string, pwHash []byte) error {
	rows, err := db.Query(`SELECT * from users WHERE email = $1`, email)
	if err != nil {
		return err
	} else if rows.Next() {
		return UserAlreadyExists
	}

	id := utils.GenRandomString(32)

	for UserIDExists(id) {
		id = utils.GenRandomString(32)
	}

	s := `INSERT INTO users
	      (id, email, pw_hash, usage, type, verified)
	      VALUES ($1, $2, $3, 0, -1, false)`

	_, err = db.Exec(s, id, email, pwHash)
	if err != nil {
		return err
	}

	return nil
}

// UserIDExists checks the users table to see if the provided id is already
// being used for another user.
func UserIDExists(id string) bool {
	rows, err := db.Query(`SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		log.Fatalf("Error querying user id: %v", err)
		return true
	}

	// If any rows are returned, the id exists
	if rows.Next() {
		return true
	}

	return false
}
