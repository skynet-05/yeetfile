package auth

import (
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/requests"
	"yeetfile/shared/endpoints"
)

func RemoveUserKeys() error {
	return config.UserConfigPaths.Reset()
}

func IsUserAuthenticated() (bool, error) {
	url := endpoints.Session.Format(config.UserConfig.Server)
	response, err := requests.GetRequest(url)
	if err != nil || response.StatusCode != http.StatusOK {
		// Ensure keys are removed
        resetErr := RemoveUserKeys()
        if resetErr != nil {
            return false, resetErr
        }

        return false, err
	}

	return true, nil
}
