package viewer

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"time"
	"unicode/utf8"
	"yeetfile/cli/commands/vault/internal"
	"yeetfile/cli/crypto"
	"yeetfile/cli/models"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared/constants"
)

type action int

const (
	PreviewFile action = iota
	Return
)

func RunViewerModel(
	item models.VaultItem,
	crypto crypto.CryptoCtx,
) (internal.Event, error) {
	var fileBytes []byte
	var err error
	if item.Size < constants.ChunkSize*3 {
		var key []byte
		_ = spinner.New().Title("Fetching file info...").Action(
			func() {
				key, err = crypto.DecryptFunc(
					crypto.DecryptionKey,
					item.ProtectedKey)
				if err != nil {
					return
				}
				fileBytes, err = downloadFile(item.ID, key)
			}).Run()
	}

	if err != nil {
		errMsg := fmt.Sprintf("Error: %s", err.Error())
		utils.ShowErrorForm(styles.ErrStyle.Render(errMsg))
		return internal.Event{
			Status: internal.StatusInvalid,
			Type:   internal.ViewFileRequest,
			Item:   item,
		}, nil
	}

	var options []huh.Option[action]
	if item.Size < constants.ChunkSize*3 {
		options = append(options, huh.NewOption("Display File", PreviewFile))
	}

	options = append(options, huh.NewOption("Return to Vault", Return))

	var selected action
	var viewerFunc func()
	viewerFunc = func() {
		err = huh.NewForm(huh.NewGroup(
			huh.NewNote().
				Title(utils.GenerateTitle(item.Name)).
				Description(
					utils.GenerateDescriptionSection(
						"Info",
						generateInfoView(item),
						21)),
			huh.NewSelect[action]().
				Options(options...).
				Value(&selected),
		)).WithTheme(styles.Theme).Run()

		if err != nil {
			return
		}

		if selected == PreviewFile {
			if fileBytes != nil {
				showFilePreview(item.Name, item.Modified, fileBytes)
				viewerFunc()
				return
			}

			_ = spinner.New().Title("Fetching file info...").Action(
				func() {
					key, err := crypto.DecryptFunc(
						crypto.DecryptionKey,
						item.ProtectedKey)
					if err != nil {
						return
					}
					fileBytes, err = downloadFile(item.ID, key)
				}).Run()

			showFilePreview(item.Name, item.Modified, fileBytes)
			viewerFunc()
			return
		}
	}

	viewerFunc()

	return internal.Event{
		Status: internal.StatusOk,
		Type:   internal.ViewFileRequest,
		Item:   item,
	}, nil
}

func showFilePreview(name string, modified time.Time, fileBytes []byte) {
	var noteContent string
	if fileBytes != nil && utf8.Valid(fileBytes) {
		showText(name, modified.Format(time.DateTime), fileBytes)
		return
	} else if fileBytes != nil && isLikelyImage(name) {
		noteContent = imageToAscii(fileBytes)
	} else if fileBytes != nil {
		noteContent = "Unable to preview file in CLI app"
	} else {
		noteContent = "File too large to preview"
	}

	_ = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(utils.GenerateTitle(name)).
				Description(noteContent),
		)).WithTheme(styles.Theme).Run()
}
