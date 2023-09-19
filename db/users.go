package db

import (
	"errors"
	"log"
	"yeetfile/utils"
)

var UserAlreadyExists = errors.New("user already exists")

// NewUser creates a new user in the "users" table, ensuring that the email
// provided is not already in use.
func NewUser(email string, pwHash []byte) (string, error) {
	rows, err := db.Query(`SELECT * from users WHERE email = $1`, email)
	if err != nil {
		return "", err
	} else if rows.Next() {
		return "", UserAlreadyExists
	}

	id := utils.GenRandomNumbers(16)

	for UserIDExists(id) {
		id = utils.GenRandomNumbers(16)
	}

	s := `INSERT INTO users
	      (id, email, pw_hash, usage, type)
	      VALUES ($1, $2, $3, 0, -1)`

	_, err = db.Exec(s, id, email, pwHash)
	if err != nil {
		return "", err
	}

	return id, nil
}

// VerifyUser uses a user's email and the token sent to their email in order
// to mark their account as verified.
func VerifyUser(email string, token string) bool {
	s := `UPDATE users
	      SET verified=true
	      WHERE email=$1 AND token=$2`

	_, err := db.Exec(s, email, token)
	if err != nil {
		panic(err)
	}

	return true
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
