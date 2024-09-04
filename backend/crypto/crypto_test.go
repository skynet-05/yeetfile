package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	text := "yeetfile"
	encryptedVal, err := Encrypt(text)
	assert.Nil(t, err)
	assert.NotEqual(t, len(text), len(encryptedVal))
	assert.NotEqual(t, text, string(encryptedVal))

	decryptedVal, err := Decrypt(encryptedVal)
	assert.Nil(t, err)

	assert.Equal(t, text, decryptedVal)
}
