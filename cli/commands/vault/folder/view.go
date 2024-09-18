package folder

import (
	"github.com/charmbracelet/huh"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/styles"
)

func RunModel() (internal.Event, error) {
	var confirmed bool
	var folderName string
	input := huh.NewInput().Title("New Folder Name").Value(&folderName)
	confirm := huh.NewConfirm().Affirmative("Create").Negative("Cancel").Value(&confirmed)

	form := huh.NewForm(huh.NewGroup(
		input,
		confirm,
	))

	err := form.WithTheme(styles.Theme).Run()
	if confirmed {
		return internal.Event{
			Value:  folderName,
			Status: internal.StatusOk,
			Type:   internal.NewFolderRequest,
		}, err
	} else {
		return internal.Event{
			Status: internal.StatusCanceled,
			Type:   internal.NewFolderRequest,
		}, err
	}
}
