package viewer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/qeesung/image2ascii/convert"
	"image"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

// Image viewer command mappings. These should always allow the following
// command format to be executed:
// <cmd> [args] path/to/file
var viewerCmdMap = map[string][]string{
	"kitty": {"+kitten", "icat"},
}

// Commands to run after image viewer command has been executed.
// These are optional if a viewer cmd doesn't have a dedicated command for
// removing images that are displayed.
var cleanupCmdMap = map[string][]string{
	"kitty": {"+kitten", "icat", "--clear"},
}

var imgExts = [...]string{
	"jpg", "jpeg", "png", "webp", "gif",
}

func isLikelyImage(name string) bool {
	for _, ext := range imgExts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	return false
}

// Checks viewerCmdMap for any matches in the user's current PATH that can
// be used to display image content on the command line.
func getImageViewerCommand() (string, []string, error) {
	for cmd, args := range viewerCmdMap {
		_, err := exec.LookPath(cmd)
		if err == nil {
			return cmd, args, nil
		}
	}

	return "", nil, errors.New("no viewer commands found")
}

// Uses a command and args (determined from getImageViewerCommand) to display
// an image in the command line.
func imageOutput(command string, args []string, fileBytes []byte) error {
	tmpFileID := shared.GenRandomString(10)
	tmpFile, err := os.CreateTemp("", tmpFileID)
	if err != nil {
		if err != nil {
			return err
		}
	}

	_, err = tmpFile.Write(fileBytes)
	if err != nil {
		return err
	}

	args = append(args, tmpFile.Name())
	err = utils.RunCmd(true, command, args...)
	if err != nil {
		return err
	}

	err = os.Remove(tmpFile.Name())
	if err != nil {
		log.Println("Unable to remove temporary file:", err)
	}

	fmt.Println("\nPress Enter key to exit image viewer")
	_, _ = fmt.Scanln()
	cleanupArgs, ok := cleanupCmdMap[command]
	if ok {
		_ = utils.RunCmd(true, command, cleanupArgs...)
	}
	_ = utils.RunCmd(true, "clear")

	return nil
}

func imageToAscii(fileBytes []byte) string {
	img, _, _ := image.Decode(bytes.NewReader(fileBytes))

	converter := convert.NewImageConverter()
	options := convert.DefaultOptions
	options.Colored = true

	imgStr := converter.Image2ASCIIString(img, &options)
	return imgStr
}

func generateInfoView(item models.VaultItem) string {
	return fmt.Sprintf("%s\n"+
		"Size: %s\n"+
		"Modified: %s\n",
		shared.EscapeString(item.Name),
		shared.ReadableFileSize(item.Size),
		item.Modified.Format(time.DateTime))
}

func downloadFile(id string, key []byte) ([]byte, error) {
	metadata, err := globals.API.GetVaultItemMetadata(id)
	if err != nil {
		return nil, err
	}

	var result []byte
	chunk := 1
	for chunk <= metadata.Chunks {
		url := endpoints.DownloadVaultFileData.Format(
			globals.Config.Server,
			metadata.ID,
			strconv.Itoa(chunk))
		chunkData, err := globals.API.DownloadFileChunk(url)
		if err != nil || len(chunkData) == 0 {
			return nil, err
		}

		decData, err := crypto.DecryptChunk(key, chunkData)
		if err != nil {
			return nil, err
		}

		result = append(result, decData...)
		chunk += 1
	}

	return result, nil
}
