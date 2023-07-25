package crypto

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
	"io"
)

const NonceSize int = 24
const KeySize int = 32

func DeriveKey(password []byte, salt []byte) ([KeySize]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, KeySize)
		if _, err := rand.Read(salt); err != nil {
			return [KeySize]byte{}, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, KeySize)
	if err != nil {
		return [KeySize]byte{}, nil, err
	}

	return [KeySize]byte(key), salt, nil
}

func EncryptChunk(key [KeySize]byte, data []byte) []byte {
	var nonce [NonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	return secretbox.Seal(nonce[:], data, &nonce, &key)
}

func GenChecksum(data []byte) ([]byte, string) {
	h := sha1.New()
	h.Write(data)

	checksum := h.Sum(nil)
	return checksum, fmt.Sprintf("%x", checksum)
}

func DecryptChunk(key [32]byte, chunk []byte) ([]byte, int, error) {
	var decryptNonce [NonceSize]byte
	copy(decryptNonce[:], chunk[:NonceSize])

	// Decrypt and append contents to output
	decrypted, ok := secretbox.Open(
		nil,
		chunk[NonceSize:],
		&decryptNonce,
		&key)

	if !ok {
		return []byte{}, 0, errors.New("failed to decrypt")
	}

	readLen := NonceSize + len(decrypted) + secretbox.Overhead
	return decrypted, readLen, nil
}

func DecryptString(key [32]byte, byteStr []byte) (string, error) {
	var decryptNonce [NonceSize]byte
	copy(decryptNonce[:], byteStr[:NonceSize])

	decrypted, ok := secretbox.Open(
		nil,
		byteStr[NonceSize:],
		&decryptNonce,
		&key)

	if !ok {
		return "", errors.New("failed to decrypt")
	}

	return string(decrypted), nil
}

func KeyFromHex(key string) [KeySize]byte {
	decodedKey, _ := hex.DecodeString(key)
	var keyBytes [KeySize]byte
	copy(keyBytes[:], decodedKey[:KeySize])

	return keyBytes
}
