package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"yeetfile/backend/config"
)

func Encrypt(text string) ([]byte, error) {
	block, err := aes.NewCipher(config.YeetFileConfig.ServerSecret)
	if err != nil {
		return nil, err
	}

	b64Text := base64.StdEncoding.EncodeToString([]byte(text))
	ciphertext := make([]byte, aes.BlockSize+len(b64Text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b64Text))
	return ciphertext, nil
}

func Decrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(config.YeetFileConfig.ServerSecret)
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)
	value, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", err
	}
	return string(value), nil
}
