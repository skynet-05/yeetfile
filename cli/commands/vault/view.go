package vault

import (
	"github.com/charmbracelet/huh"
	"yeetfile/cli/commands/vault/confirmation"
	"yeetfile/cli/commands/vault/filepicker"
	"yeetfile/cli/commands/vault/files"
	"yeetfile/cli/commands/vault/folder"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/commands/vault/rename"
	"yeetfile/cli/commands/vault/share"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
)

func ShowVaultModel() {
	m, err := files.RunFilesModel(files.Model{}, internal.Event{})
	for err == nil && m.ViewRequest.View > internal.NullView {
		if err != nil {
			utils.HandleCLIError("Error in vault view", err)
			return
		}

		var event internal.Event
		var subviewErr error
		switch m.ViewRequest.View {
		case internal.FilePickerView:
			event, subviewErr = filepicker.RunModel()
		case internal.ConfirmationView:
			event, subviewErr = confirmation.RunModel(
				m.ViewRequest.Type,
				m.ViewRequest.Item)
		case internal.NewFolderView:
			event, subviewErr = folder.RunModel()
		case internal.RenameView:
			event, subviewErr = rename.RunModel(m.ViewRequest.Item)
		case internal.ShareView:
			event, subviewErr = share.RunModel(
				m.ViewRequest.Item,
				nil,
				m.Context.Crypto.DecryptFunc,
				m.Context.Crypto.DecryptionKey)
		case internal.FilesView:
			m, err = files.RunFilesModel(m, m.IncomingEvent)
			continue
		}

		utils.HandleCLIError("Error in subview", subviewErr)
		m, err = files.RunFilesModel(m, event)
	}
}

func ShowVaultPasswordPromptModel(errorMsgs ...string) []byte {
	var password string
	desc := "Enter your vault session password below to continue"
	if len(errorMsgs) > 0 {
		desc = errorMsgs[0]
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(utils.GenerateTitle(
				"Vault Session Password")).
				Description(desc),
			huh.NewInput().Title("Password").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewConfirm().Affirmative("Submit").Negative(""),
		),
	).WithTheme(styles.Theme).Run()

	utils.HandleCLIError("error showing vault pw prompt", err)

	return []byte(password)
}
