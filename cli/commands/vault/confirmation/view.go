package confirmation

import (
	"github.com/charmbracelet/huh"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
)

func RunModel(req internal.RequestType, item models.VaultItem) (internal.Event, error) {
	var confirmed bool
	confirm := huh.NewConfirm().Affirmative("Yes").Negative("No").Value(&confirmed)

	title, desc := GenConfirmMsg(req, item)
	confirm.Title(title)
	confirm.Description(desc)

	theme := styles.Theme
	if req == internal.DeleteFileRequest {
		theme = styles.DestructiveTheme()
	}

	err := huh.NewForm(huh.NewGroup(confirm)).WithTheme(theme).Run()
	if confirmed {
		return internal.Event{
			Status: internal.StatusOk,
			Type:   req,
			Item:   item,
		}, err
	} else {
		return internal.Event{
			Status: internal.StatusCanceled,
		}, err
	}
}
