package crypto

import (
	"bytes"
	"golang.org/x/crypto/nacl/secretbox"
	"testing"
	"yeetfile/shared"
)

var data = []byte("data")
var password = []byte("topsecret")

func TestDeriveKey(t *testing.T) {
	key, salt, pepper, err := DeriveKey(password, nil, nil)
	if err != nil {
		t.Fatalf("Error generating key: %v\n", err)
	}

	if len(salt) == 0 || len(pepper) == 0 {
		t.Fatalf("Failed to generate salt or pepper")
	}

	isEmpty := true
	for _, b := range key {
		if b != 0 {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		t.Fatalf("Generated key is empty")
	}
}

func TestEncryptChunk(t *testing.T) {
	key, _, _, _ := DeriveKey(password, nil, nil)
	encrypted := EncryptChunk(key, data)

	if len(encrypted) != len(data)+shared.NonceSize+secretbox.Overhead {
		t.Fatalf("Unexpected encrypted data size\n" +
			"(Should be data length + nonce_size + overhead)")
	}
}

func TestDecryptChunk(t *testing.T) {
	key, salt, pepper, _ := DeriveKey(password, nil, nil)
	encrypted := EncryptChunk(key, data)

	decryptKey, _, _, _ := DeriveKey(password, salt, pepper)
	decrypted, err := DecryptChunk(decryptKey, encrypted)

	if err != nil {
		t.Fatalf("Error decrypting data: %v\n", err)
	} else if !bytes.Equal(decrypted, data) {
		t.Fatalf("Decrypted data doesn't match source data")
	}
}
