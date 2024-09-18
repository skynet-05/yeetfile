package vault

import (
	"yeetfile/cli/commands/vault/confirmation"
	"yeetfile/cli/commands/vault/filepicker"
	"yeetfile/cli/commands/vault/files"
	"yeetfile/cli/commands/vault/folder"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/commands/vault/rename"
	"yeetfile/cli/commands/vault/share"
	"yeetfile/cli/commands/vault/viewer"
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
		case internal.FileViewerView:
			event, subviewErr = viewer.RunViewerModel(
				m.ViewRequest.Item,
				m.ViewRequest.CryptoCtx)
		case internal.FilesView:
			m, err = files.RunFilesModel(m, m.IncomingEvent)
			continue
		}

		utils.HandleCLIError("Error in subview", subviewErr)
		m, err = files.RunFilesModel(m, event)
	}
}
