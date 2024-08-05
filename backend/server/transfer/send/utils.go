package send

import (
	"errors"
	"log"
	"net/http"
	"os"
	"yeetfile/backend/db"
	"yeetfile/backend/server/session"
)

var OutOfSpaceError = errors.New("not enough space to upload")

// UserCanSend fetches the user ID associated with the request and checks to
// see if they have enough remaining send space to send a file
func UserCanSend(size int, req *http.Request) (bool, error) {
	if os.Getenv("YEETFILE_DEBUG") == "1" {
		return true, nil
	}

	// Validate that the user has enough space to upload this file
	s, err := session.GetSession(req)
	if err != nil {
		return false, err
	}

	id := session.GetSessionUserID(s)
	usedSend, availableSend, err := db.GetUserSendLimits(id)
	if err != nil {
		log.Printf("Error validating ability to upload: %v\n", err)
		return false, err
	} else if availableSend-usedSend < size {
		return false, OutOfSpaceError
	}

	return true, nil
}

// UpdateUserMeter receives the size of an uploaded chunk and subtracts that
// value from the user's available storage meter
func UpdateUserMeter(size int, req *http.Request) error {
	s, err := session.GetSession(req)
	if err != nil {
		return err
	}

	id := session.GetSessionUserID(s)
	err = db.UpdateUserSendUsed(id, size)
	if err != nil {
		log.Printf("Failed to reduce user storage: %v\n", err)
		return err
	}

	return nil
}
