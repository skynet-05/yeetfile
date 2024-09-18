package viewer

import (
	"bytes"
	"fmt"
	"github.com/qeesung/image2ascii/convert"
	"image"
	"strconv"
	"strings"
	"time"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/models"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

var imgExts = [...]string{
	"jpg", "jpeg", "png",
}

func isLikelyImage(name string) bool {
	for _, ext := range imgExts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	return false
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
		if err != nil {
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
