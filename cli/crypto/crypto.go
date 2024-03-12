package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
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
	key := pbkdf2.Key(pepperPw, salt, 100000, shared.KeySize, sha256.New)

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

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	result := aesgcm.Seal(nil, nonce[:], data, nil)
	var merged []byte
	merged = append(merged, nonce[:]...)
	merged = append(merged, result[:]...)

	return merged
}

// DecryptChunk decrypts an encrypted chunk of data using the provided key. If
// the key is unable to decrypt the data, an error is returned, otherwise the
// decrypted data is returned.
func DecryptChunk(key [shared.KeySize]byte, chunk []byte) ([]byte, error) {
	nonce := chunk[:shared.NonceSize]
	data := chunk[shared.NonceSize:]

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	//readLen := shared.NonceSize + len(decrypted) + secretbox.Overhead
	return plaintext, nil
}

// DecryptString decrypts a string using DecryptChunk, but returns a string
// directly rather than returning a byte slice
func DecryptString(key [shared.KeySize]byte, byteStr []byte) (string, error) {
	decrypted, err := DecryptChunk(key, byteStr)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
