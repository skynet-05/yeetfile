package download

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"io"
	"net/http"
	"strings"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/cli/requests"
	"yeetfile/cli/styles"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

var decryptError = errors.New("decryption error")

type DownloadResource struct {
	Server string
	ItemID string
	Pepper string
}

type PreparedDownload struct {
	ID         string
	Server     string
	Name       string
	Size       int
	Chunks     int
	Key        []byte
	Expiration time.Time
	Downloads  int
	IsText     bool
}

func parseLink(link string) DownloadResource {
	linkSegments := strings.Split(link, "/")
	server := strings.Join(linkSegments[0:len(linkSegments)-1], "/")
	resource := strings.Split(linkSegments[len(linkSegments)-1], "#")

	var id string
	var pepper string
	if len(resource) == 1 {
		id = resource[0]
	} else if len(resource) > 1 {
		id = resource[0]
		pepper = resource[1]
	}

	return DownloadResource{
		Server: server,
		ItemID: id,
		Pepper: pepper,
	}
}

func prepDownload(link string) (PreparedDownload, error) {
	d := parseLink(link)
	metadata, err := d.fetchMetadata()
	if err != nil {
		return PreparedDownload{}, err
	}

	name, key, err := d.decryptResponse(metadata, "")
	if err != nil {
		return PreparedDownload{}, err
	}

	prep := PreparedDownload{
		ID:         metadata.ID,
		Name:       name,
		Key:        key,
		Size:       metadata.Size,
		Chunks:     metadata.Chunks,
		Expiration: metadata.Expiration,
		Downloads:  metadata.Downloads,
		Server:     d.Server,
		IsText:     strings.HasPrefix(metadata.ID, constants.PlaintextIDPrefix),
	}

	return prep, nil
}

func (d DownloadResource) fetchMetadata() (shared.DownloadResponse, error) {
	url := endpoints.DownloadSendFileMetadata.Format(d.Server, d.ItemID)
	resp, err := requests.GetRequest(url)
	if err != nil {
		return shared.DownloadResponse{}, err
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("server error [%d]: %s", resp.StatusCode, body)
		return shared.DownloadResponse{}, errors.New(msg)
	}

	var downloadResponse shared.DownloadResponse
	err = json.NewDecoder(resp.Body).Decode(&downloadResponse)
	if err != nil {
		return shared.DownloadResponse{}, err
	}

	return downloadResponse, nil
}

func getPassword(err error) (string, error) {
	desc := "This content is password protected"
	if err != nil {
		desc = styles.ErrStyle.Render(err.Error())
	}

	var password string
	pErr := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Password").
			Description(desc).
			EchoMode(huh.EchoModePassword).
			Value(&password),
		huh.NewConfirm().Affirmative("Submit").Negative(""),
	)).WithTheme(styles.Theme).Run()

	return password, pErr
}

// decryptResponse decrypts the name from the shared.DownloadResponse returned
// by the server, and returns the decrypted name and the key that was successfully
// used to decrypt the name.
func (d DownloadResource) decryptResponse(
	response shared.DownloadResponse,
	password string,
) (string, []byte, error) {
	key, _, _, err := crypto.DeriveSendingKey(
		[]byte(password),
		response.Salt,
		[]byte(d.Pepper))

	encName, err := hex.DecodeString(response.Name)
	if err != nil {
		return "", nil, err
	}

	decName, err := crypto.DecryptChunk(key, encName)
	if err != nil {
		// Decryption error, likely means the file is password protected
		var newPassword string
		if password == "" {
			newPassword, err = getPassword(nil)
		} else {
			newPassword, err = getPassword(decryptError)
		}

		if err != nil {
			return "", nil, err
		}

		return d.decryptResponse(response, newPassword)
	}

	return string(decName), key, nil
}

func generateDescription(download PreparedDownload) string {
	name := download.Name
	if strings.HasPrefix(download.ID, constants.PlaintextIDPrefix) {
		name = "N/A (text-only)"
	}

	timeDiff := download.Expiration.Sub(time.Now())

	return fmt.Sprintf(""+
		"- Name: %s\n"+
		"- Size: %s\n"+
		"- Expiration: %s (%s)\n"+
		"- Downloads Remaining: %d\n",
		name,
		shared.ReadableFileSize(download.Size),
		utils.LocalTimeFromUTC(download.Expiration),
		timeDiff,
		download.Downloads,
	)
}
