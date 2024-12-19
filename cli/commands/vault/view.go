package vault

import (
	"log"
	"yeetfile/cli/commands/vault/confirmation"
	"yeetfile/cli/commands/vault/filepicker"
	"yeetfile/cli/commands/vault/folder"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/commands/vault/items"
	"yeetfile/cli/commands/vault/pass"
	"yeetfile/cli/commands/vault/rename"
	"yeetfile/cli/commands/vault/share"
	"yeetfile/cli/commands/vault/viewer"
	"yeetfile/cli/utils"
)

func ShowPassVaultModel() {
	m, err := items.RunVaultModel(
		items.Model{IsPassVault: true},
		internal.Event{})
	if err != nil {
		log.Fatal(err)
	}

	showVaultModel(m)
}

func ShowFileVaultModel() {
	m, err := items.RunVaultModel(
		items.Model{IsPassVault: false},
		internal.Event{})
	if err != nil {
		log.Fatal(err)
	}

	showVaultModel(m)
}

func showVaultModel(m items.Model) {
	var err error
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
		case internal.FileViewerView:
			event, subviewErr = viewer.RunViewerModel(
				m.ViewRequest.Item,
				m.ViewRequest.CryptoCtx)
		case internal.NewPassView:
			event, subviewErr = pass.RunNewPassEntryModel()
		case internal.EditPassView:
			event, subviewErr = pass.RunEditPassEntryModel(m.ViewRequest.Item)
		case internal.ViewPassView:
			subviewErr = pass.RunViewPassEntryModel(m.ViewRequest.Item)
		case internal.FilesView:
			m, err = items.RunVaultModel(m, m.IncomingEvent)
			continue
		}

		utils.HandleCLIError("Error in subview", subviewErr)
		m, err = items.RunVaultModel(m, event)
	}
}
