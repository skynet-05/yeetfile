package db

import (
	"errors"
	"fmt"
	"log"
	"yeetfile/utils"
)

var defaultUsage = 1024 * 1024 * 10 // 10mb

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
	paymentID := utils.GenRandomString(16)

	for UserIDExists(id) {
		id = utils.GenRandomNumbers(16)
	}

	s := `INSERT INTO users
	      (id, email, pw_hash, usage, payment_id)
	      VALUES ($1, $2, $3, $4, $5)`

	_, err = db.Exec(s, id, email, pwHash, defaultUsage, paymentID)
	if err != nil {
		return "", err
	}

	return id, nil
}

// RotateUserPaymentID overwrites the previous payment ID once a transaction is
// completed and storage has been added to their account.
func RotateUserPaymentID(paymentID string) error {
	rows, err := db.Query(`SELECT id from users WHERE payment_id = $1`, paymentID)
	if err != nil {
		return err
	} else if !rows.Next() {
		errorStr := fmt.Sprintf("unable to find user with payment id '%s'", paymentID)
		return errors.New(errorStr)
	}

	newID := utils.GenRandomString(16)
	for PaymentIDExists(newID) {
		newID = utils.GenRandomString(16)
	}

	// Read in account ID for the user
	var accountID string
	err = rows.Scan(&accountID)

	// Replace payment ID
	s := `UPDATE users
	      SET payment_id=$1
	      WHERE id=$2`

	_, err = db.Exec(s, newID, accountID)
	if err != nil {
		return err
	}

	return nil

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

// PaymentIDExists checks the user table to see if the provided payment ID
// (for Stripe + BTCPay) already exists for another user.
func PaymentIDExists(paymentID string) bool {
	rows, err := db.Query(`SELECT * FROM users WHERE payment_id = $1`, paymentID)
	if err != nil {
		log.Fatalf("Error querying user payment id: %v", err)
		return true
	}

	// If any rows are returned, the id exists
	if rows.Next() {
		return true
	}

	return false
}

// AddUserStorage adds amount to the usage column for a user with the matching
// payment ID. Once the payment ID is used here, it should be replaced by calling
// RotateUserPaymentID.
func AddUserStorage(paymentID string, amount int) error {
	s := `UPDATE users
	      SET usage=usage + $1
	      WHERE payment_id=$2`

	_, err := db.Exec(s, amount, paymentID)
	if err != nil {
		return err
	}

	return nil
}
