package logout

import (
	"errors"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/requests"
	"yeetfile/shared/endpoints"
)

func LogOut() error {
	url := endpoints.Logout.Format(config.UserConfig.Server)
	response, err := requests.GetRequest(url)
	if err != nil {
		return err
	} else if response.StatusCode >= http.StatusBadRequest {
		return errors.New("error logging out")
	}

	err = config.UserConfigPaths.Reset()
	return err
}
