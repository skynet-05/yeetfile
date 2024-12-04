package rename

import (
	"github.com/charmbracelet/huh"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
)

func RunModel(item models.VaultItem) (internal.Event, error) {
	newName := item.Name
	var confirmed bool

	title := "Rename File"
	if item.IsFolder {
		title = "Rename Folder"
	}

	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(title).
			Description("Enter a new name below").
			Placeholder(item.Name).
			Value(&newName),
		huh.NewConfirm().
			Affirmative("Rename").
			Negative("Cancel").
			Value(&confirmed),
	)).WithTheme(styles.Theme).Run()

	if confirmed {
		return internal.Event{
			Value:  newName,
			Status: internal.StatusOk,
			Type:   internal.RenameRequest,
			Item:   item,
		}, err
	} else {
		return internal.Event{
			Status: internal.StatusCanceled,
			Type:   internal.RenameRequest,
		}, err
	}
}
