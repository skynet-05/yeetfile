package main

import (
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/config"
)

func SetSessionFromCookies(response *http.Response) error {
	// User signed up for an account ID only, and should be logged
	// in now
	cookies := response.Cookies()
	if len(cookies) > 0 {
		session = cookies[0].Value
		err := config.SetSession(configPaths, session)
		if err != nil {
			fmt.Printf("Failed to save user session: %v\n", err)
			return err
		}
	} else {
		return errors.New("failed to save new session")
	}

	return nil
}
