package logout

import (
	"fmt"
	"github.com/charmbracelet/huh/spinner"
	"yeetfile/cli/utils"
)

func ShowLogoutModel() {
	_ = spinner.New().Title("Logging out...").Action(
		func() {
			err := LogOut()
			utils.HandleCLIError("error logging out", err)
		}).Run()

	fmt.Println("You are logged out")
}
