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
	key, _, _, _ := DeriveSendingKey(password, nil, nil)
	encrypted := EncryptChunk(key, data)

	if len(encrypted) != len(data)+shared.TotalOverhead {
		t.Fatalf("Unexpected encrypted data size\n"+
			"expected: %d, actual: %d",
			len(data)+shared.TotalOverhead,
			len(encrypted))
	}

	storageKey, _ := CreateStorageKey()
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
