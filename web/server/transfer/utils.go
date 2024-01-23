package transfer

import (
	"log"
	"net/http"
	"yeetfile/web/db"
	"yeetfile/web/server/session"
)

// UserCanUpload fetches the user ID associated with the request and checks to
// see if their upload meter is within 1 chunk size of the data being uploaded
func UserCanUpload(size int, req *http.Request) bool {
	// Validate that the user has enough space to upload this file
	s, err := session.GetSession(req)
	if err != nil {
		return false
	}

	id := session.GetSessionUserID(s)
	meter, err := db.GetUserMeter(id)
	if err != nil || meter < size {
		return false
	}

	return true
}

// UpdateUserMeter receives the size of an uploaded chunk and subtracts that
// value from the user's available storage meter
func UpdateUserMeter(size int, req *http.Request) error {
	s, err := session.GetSession(req)
	if err != nil {
		return err
	}

	id := session.GetSessionUserID(s)
	err = db.ReduceUserStorage(id, size)
	if err != nil {
		log.Printf("Failed to reduce user storage: %v\n", err)
		return err
	}

	return nil
}
