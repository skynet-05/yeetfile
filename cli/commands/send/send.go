package send

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/transfer"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

type fileUpload struct {
	FilePath     string
	MaxDownloads int
	ExpUnits     string
	ExpValue     int
	Password     string
}

type textUpload struct {
	Text         string
	MaxDownloads int
	ExpUnits     string
	ExpValue     int
	Password     string
}

const (
	expMinutes = "minutes"
	expHours   = "hours"
	expDays    = "days"
)

func getDuration(value int64, units string) time.Duration {
	var duration time.Duration
	switch units {
	case expMinutes:
		duration = time.Duration(value * int64(time.Minute))
	case expHours:
		duration = time.Duration(value * int64(time.Hour))
	case expDays:
		duration = time.Duration(value * int64(time.Hour*24))
	}

	return duration
}

func getExpString(value int64, units string) string {
	duration := getDuration(value, units)
	return time.Now().Add(duration).Format("02 Jan 2006 15:04 MST")
}

func isValidExp(value int64, units string) bool {
	duration := getDuration(value, units)
	maxAge := time.Now().Add(constants.MaxSendAgeDays * time.Hour * 24)
	return time.Now().Add(duration).Before(maxAge)
}

func createTextLink(upload textUpload) (string, string, error) {
	key, salt, pepper, err := crypto.DeriveSendingKey(
		[]byte(upload.Password), nil, nil)
	if err != nil {
		return "", "", err
	}

	encName, err := crypto.EncryptChunk(key, []byte(shared.GenRandomString(8)))
	if err != nil {
		return "", "", err
	}
	hexEncName := hex.EncodeToString(encName)
	encText, err := crypto.EncryptChunk(key, []byte(upload.Text))
	if err != nil {
		return "", "", err
	}

	encTextUpload := shared.PlaintextUpload{
		Name:       hexEncName,
		Salt:       salt,
		Downloads:  upload.MaxDownloads,
		Expiration: createExpString(upload.ExpValue, upload.ExpUnits),
		Text:       encText,
	}

	id, err := globals.API.UploadText(encTextUpload)
	if err != nil {
		return "", "", err
	}

	return id, string(pepper), nil
}

func createFileLink(upload fileUpload, progress func(int, int)) (string, string, error) {
	key, salt, pepper, err := crypto.DeriveSendingKey(
		[]byte(upload.Password), nil, nil)
	if err != nil {
		return "", "", err
	}

	file, stat, err := shared.GetFileInfo(upload.FilePath)

	encName, err := crypto.EncryptChunk(key, []byte(stat.Name()))
	hexEncName := hex.EncodeToString(encName)
	size := int(stat.Size())
	numChunks := transfer.GetNumChunks(stat.Size())

	metadata := shared.UploadMetadata{
		Name:       hexEncName,
		Chunks:     numChunks,
		Size:       size,
		Salt:       salt,
		Downloads:  upload.MaxDownloads,
		Expiration: createExpString(upload.ExpValue, upload.ExpUnits),
	}

	pending, err := transfer.InitSendFile(file, metadata, key)
	if err != nil {
		return "", "", err
	}

	chunk := 0
	result, err := pending.UploadData(func() {
		chunk += 1
		progress(chunk, pending.NumChunks)
	})
	if err != nil {
		return "", "", err
	}

	return result, string(pepper), nil
}

func createExpString(expValue int, expUnits string) string {
	return fmt.Sprintf("%d%s", expValue, strings.ToLower(string(expUnits[0])))
}
