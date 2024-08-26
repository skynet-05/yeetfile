package crypto

import (
	"bytes"
	"testing"
	"yeetfile/shared/constants"
)

var data = []byte("data")
var password = []byte("topsecret")

func TestDeriveKey(t *testing.T) {
	key, salt, err := DeriveSendingKey(password, nil)
	if err != nil {
		t.Fatalf("Error generating key: %v\n", err)
	}

	if len(salt) == 0 {
		t.Fatalf("Failed to generate salt")
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
	plainData := make([]byte, constants.ChunkSize)
	key, _, _ := DeriveSendingKey(password, nil)
	encrypted, _ := EncryptChunk(key, plainData)

	if len(encrypted) != len(plainData)+constants.TotalOverhead {
		t.Fatalf("Unexpected encrypted data size\n"+
			"expected: %d, actual: %d",
			len(plainData)+constants.TotalOverhead,
			len(encrypted))
	}

	storageKey, _ := GenerateRandomKey()
	storageEncrypted, _ := EncryptChunk(storageKey, data)

	if len(storageEncrypted) != len(data)+constants.TotalOverhead {
		t.Fatalf("Unexpected encrypted storage data size\n"+
			"expected: %d, actual: %d",
			len(data)+constants.TotalOverhead,
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
	key, salt, _ := DeriveSendingKey(password, nil)
	encrypted, _ := EncryptChunk(key, data)

	decryptKey, _, _ := DeriveSendingKey(password, salt)
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

	if len(loginKey) != len(newLoginKey) {
		t.Fatalf("Login key hash lengths do not match")
	} else {
		for i, b := range loginKey {
			if b != newLoginKey[i] {
				t.Fatal("Login key hash contents don't match")
			}
		}
	}
}
