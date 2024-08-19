package auth

import (
	"golang.org/x/crypto/bcrypt"
	"strings"
	"yeetfile/backend/db"
)

// ValidateCredentials checks the provided key hash against the one stored in
// the database, and if there's a match, returns the user's true account ID.
func ValidateCredentials(identifier string, keyHash []byte) (string, error) {
	var userID string
	var pwHash []byte
	var err error
	if strings.Contains(identifier, "@") {
		pwHash, err = db.GetUserPasswordHashByEmail(identifier)
		if err != nil {
			return "", err
		}

		userID, err = db.GetUserIDByEmail(identifier)
		if err != nil {
			return "", err
		}
	} else {
		pwHash, err = db.GetUserPasswordHashByID(identifier)
		if err != nil {
			return "", err
		}

		userID = identifier
	}

	err = bcrypt.CompareHashAndPassword(pwHash, keyHash)
	if err != nil {
		return "", err
	}

	return userID, nil
}
