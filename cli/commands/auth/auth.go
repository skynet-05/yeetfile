package auth

import (
	"yeetfile/cli/globals"
)

func IsUserAuthenticated() (bool, error) {
	_, err := globals.API.GetSession()
	if err != nil {
		// Ensure keys are removed
		resetErr := globals.Config.Reset()
		if resetErr != nil {
			return false, resetErr
		}

		return false, err
	}

	return true, nil
}
