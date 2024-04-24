package crypto

import (
	"bytes"
	"testing"
	"yeetfile/shared"
)

var data = []byte("data")
var password = []byte("topsecret")

func TestDeriveKey(t *testing.T) {
	key, salt, pepper, err := DeriveSendingKey(password, nil, nil)
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
	plainData := make([]byte, shared.ChunkSize)
	key, _, _, _ := DeriveSendingKey(password, nil, nil)
	encrypted := EncryptChunk(key, plainData)

	if len(encrypted) != len(plainData)+shared.TotalOverhead {
		t.Fatalf("Unexpected encrypted data size\n"+
			"expected: %d, actual: %d",
			len(plainData)+shared.TotalOverhead,
			len(encrypted))
	}

	storageKey, _ := GenerateStorageKey()
	storageEncrypted := EncryptChunk(storageKey, data)

	if len(storageEncrypted) != len(data)+shared.TotalOverhead {
		t.Fatalf("Unexpected encrypted storage data size\n"+
			"expected: %d, actual: %d",
			len(data)+shared.TotalOverhead,
			len(storageEncrypted))
	} else {
		equal := true
		for i := range storageEncrypted {
			if storageEncrypted[i] != encrypted[i] {
				equal = false
			}
		}

		if equal {
			t.Fatalf("Sending and storage encrypted files are the " +
				"same, they should be different")
		}
	}
}

func TestDecryptChunk(t *testing.T) {
	key, salt, pepper, _ := DeriveSendingKey(password, nil, nil)
	encrypted := EncryptChunk(key, data)

	decryptKey, _, _, _ := DeriveSendingKey(password, salt, pepper)
	decrypted, err := DecryptChunk(decryptKey, encrypted)

	if err != nil {
		t.Fatalf("Error decrypting data: %v\n", err)
	} else if !bytes.Equal(decrypted, data) {
		t.Fatalf("Decrypted data doesn't match source data")
	}
}

func TestLoginKey(t *testing.T) {
	myPassword := []byte("my-password")
	myEmail := []byte("myemail@domain.com")

	storageKey := GenerateUserKey(myEmail, myPassword)
	loginKey := GenerateLoginKeyHash(storageKey, myPassword)

	// Simulates login at a later time
	newStorageKey := GenerateUserKey(myEmail, myPassword)
	newLoginKey := GenerateLoginKeyHash(newStorageKey, myPassword)

	if loginKey != newLoginKey {
		t.Fatalf("Login key hashes do not match")
	}
}
