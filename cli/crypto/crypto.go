package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"log"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

// DeriveSendingKey uses PBKDF2 to derive a key for sending a file. Both the
// salt and the pepper can be left nil in order to randomly generate both values.
func DeriveSendingKey(
	password []byte,
	salt []byte,
	pepper []byte,
) ([]byte, []byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, shared.KeySize)
		if _, err := rand.Read(salt); err != nil {
			return []byte{}, nil, nil, err
		}
	}

	if pepper == nil {
		pepper = []byte(utils.GeneratePassphrase())
	}

	pepperPw := append(password, pepper...)
	key := DerivePBKDFKey(pepperPw, salt)

	return key, salt, pepper, nil
}

// GenerateStorageKey creates the 256-bit symmetric key used for encrypting files
// that are stored (not sent) in YeetFile. This is always encrypted using the
// master PBKDF2 key before being sent to the server.
func GenerateStorageKey() ([]byte, error) {
	key := make([]byte, shared.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

// DerivePBKDFKey uses PBKDF2 to derive a key from a known password and salt.
func DerivePBKDFKey(password []byte, salt []byte) []byte {
	key := pbkdf2.Key(password, salt, 600000, shared.KeySize, sha256.New)
	return key
}

// GenerateUserKey generates the key used for encrypting and decrypting
// files that are stored in YeetFile, using their identifier (email or acct ID)
// and their password.
func GenerateUserKey(identifier []byte, password []byte) []byte {
	return DerivePBKDFKey(password, identifier)
}

// GenerateLoginKeyHash generates a login key using the user's user key and
// their password, and returns a hex encoded hash of the resulting key.
func GenerateLoginKeyHash(userKey []byte, password []byte) string {
	loginKey := DerivePBKDFKey(userKey, password)

	h := sha256.New()
	h.Write(loginKey)
	loginKeyHash := h.Sum(nil)

	return hex.EncodeToString(loginKeyHash)
}

// EncryptChunk encrypts a chunk of data using either the sending or storage key.
// Returns the encrypted chunk of data.
func EncryptChunk(key []byte, data []byte) []byte {
	var iv [shared.IVSize]byte
	if _, err := io.ReadFull(rand.Reader, iv[:]); err != nil {
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

	result := aesgcm.Seal(nil, iv[:], data, nil)
	var merged []byte
	merged = append(merged, iv[:]...)
	merged = append(merged, result[:]...)

	return merged
}

// DecryptChunk decrypts an encrypted chunk of data using the provided key. If
// the key is unable to decrypt the data, an error is returned, otherwise the
// decrypted data is returned.
func DecryptChunk(key []byte, chunk []byte) ([]byte, error) {
	iv := chunk[:shared.IVSize]
	data := chunk[shared.IVSize:]

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, iv, data, nil)
	if err != nil {
		return nil, err
	}

	//readLen := shared.NonceSize + len(decrypted) + secretbox.Overhead
	return plaintext, nil
}

// DecryptString decrypts a string using DecryptChunk, but returns a string
// directly rather than returning a byte slice
func DecryptString(key []byte, byteStr []byte) (string, error) {
	decrypted, err := DecryptChunk(key, byteStr)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
