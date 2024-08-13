package logout

import (
	"yeetfile/cli/globals"
)

func LogOut() error {
	err := globals.API.LogOut()
	if err != nil {
		return err
	}

	err = globals.Config.Reset()
	return err
}
