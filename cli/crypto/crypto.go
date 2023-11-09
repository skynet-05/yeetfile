package crypto

import (
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
	"io"
	"log"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

// DeriveKey derives a key from a password, salt, and pepper. Both the salt and
// the pepper can be left nil in order to randomly generate both values.
func DeriveKey(
	password []byte,
	salt []byte,
	pepper []byte,
) ([shared.KeySize]byte, []byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, shared.KeySize)
		if _, err := rand.Read(salt); err != nil {
			return [shared.KeySize]byte{}, nil, nil, err
		}
	}

	if pepper == nil {
		pepper = []byte(utils.GeneratePassphrase())
	}

	pepperPw := append(password, pepper...)

	key, err := scrypt.Key(pepperPw, salt, 32768, 8, 1, shared.KeySize)
	if err != nil {
		return [shared.KeySize]byte{}, nil, nil, err
	}

	var keyOut [shared.KeySize]byte
	copy(keyOut[:], key)

	return keyOut, salt, pepper, nil
}

// EncryptChunk encrypts a chunk of data using a key from DeriveKey. Returns the
// encrypted chunk of data.
func EncryptChunk(key [shared.KeySize]byte, data []byte) []byte {
	var nonce [shared.NonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		log.Fatalf("Error generating nonce: %v\n", err)
	}

	return secretbox.Seal(nonce[:], data, &nonce, &key)
}

// DecryptChunk decrypts an encrypted chunk of data using the provided key. If
// the key is unable to decrypt the data, an error is returned, otherwise the
// decrypted data is returned.
func DecryptChunk(key [32]byte, chunk []byte) ([]byte, error) {
	var decryptNonce [shared.NonceSize]byte
	copy(decryptNonce[:], chunk[:shared.NonceSize])

	// Decrypt and append contents to output
	decrypted, ok := secretbox.Open(
		nil,
		chunk[shared.NonceSize:],
		&decryptNonce,
		&key)

	if !ok {
		return []byte{}, errors.New("failed to decrypt")
	}

	//readLen := shared.NonceSize + len(decrypted) + secretbox.Overhead
	return decrypted, nil
}

// DecryptString decrypts a string using DecryptChunk, but returns a string
// directly rather than returning a byte slice
func DecryptString(key [32]byte, byteStr []byte) (string, error) {
	decrypted, err := DecryptChunk(key, byteStr)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
