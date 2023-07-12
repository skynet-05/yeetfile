package crypto

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
	"io"
	"log"
	"os"
)

// const BUFFER_SIZE int = 10485760 // 10mb (b2 min part is 5mb)
const BUFFER_SIZE int = 5242880
const NONCE_SIZE int = 24
const KEY_SIZE int = 32

func TestEncryptAndDecrypt() {
	filename := "lipsum.txt"
	password := []byte("topsecret")

	key, salt, err := DeriveKey(password, nil)
	if err != nil {
		log.Fatalf("Failed to derive key: %v", err.Error())
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		panic("Unable to open file")
	}

	output, err := os.OpenFile("test.enc", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		panic("Unable to open output file")
	}

	idx := 0
	for idx < len(file) {
		chunkSize := BUFFER_SIZE
		if idx+BUFFER_SIZE > len(file) {
			chunkSize = len(file) - idx
		}

		chunk := EncryptChunk(key, file[idx:idx+chunkSize])
		_, checksum := GenChecksum(chunk)

		fmt.Println(fmt.Sprintf("Chunk checksum: %x", checksum))

		_, err = output.Write(chunk)
		if err != nil {
			panic("Failed to write encrypted chunk to the output file")
		}

		idx += chunkSize
	}

	_, err = output.Write(salt)
	if err != nil {
		panic("Failed to write salt to the output file")
	}

	err = output.Close()
	if err != nil {
		panic("Failed to close the output file")
	}

	out, err := os.ReadFile("out.enc")
	if err != nil {
		panic("Failed to read the output file")
	}

	plaintext := OldDecrypt(password, out)

	fmt.Println(string(plaintext))
}

func DeriveKey(password []byte, salt []byte) ([KEY_SIZE]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, KEY_SIZE)
		if _, err := rand.Read(salt); err != nil {
			return [KEY_SIZE]byte{}, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, KEY_SIZE)
	if err != nil {
		return [KEY_SIZE]byte{}, nil, err
	}

	return [KEY_SIZE]byte(key), salt, nil
}

func EncryptChunk(key [KEY_SIZE]byte, data []byte) []byte {
	var nonce [NONCE_SIZE]byte
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

func DecryptChunk(key [32]byte, chunk []byte) ([]byte, int) {
	var decryptNonce [NONCE_SIZE]byte
	copy(decryptNonce[:], chunk[:NONCE_SIZE])

	// Define the size of chunk to read
	//chunkSize := NONCE_SIZE + BUFFER_SIZE + secretbox.Overhead
	//if chunkSize > len(chunk)+NONCE_SIZE {
	//	// Read remainder of file if the chunk size exceeds the available data
	//	chunkSize = len(chunk)
	//}

	// Decrypt and append contents to output
	decrypted, ok := secretbox.Open(nil, chunk[NONCE_SIZE:], &decryptNonce, &key)
	if !ok {
		panic("decryption error")
	}

	return decrypted, NONCE_SIZE + len(decrypted) + secretbox.Overhead
}

func Decrypt(password []byte, salt []byte, data []byte) []byte {
	key, _, err := DeriveKey(password, salt)
	if err != nil {
		return nil
	}

	var output []byte
	for len(data) > 0 {
		var decryptNonce [NONCE_SIZE]byte
		copy(decryptNonce[:], data[:NONCE_SIZE])

		// Define the size of chunk to read
		chunkSize := NONCE_SIZE + BUFFER_SIZE + secretbox.Overhead
		if chunkSize > len(data)+NONCE_SIZE {
			// Read remainder of file if the chunk size exceeds the available data
			chunkSize = len(data)
		}

		// Decrypt and append contents to output
		decrypted, ok := secretbox.Open(nil, data[NONCE_SIZE:chunkSize], &decryptNonce, &key)
		if !ok {
			panic("decryption error")
		}

		output = append(output, decrypted...)
		data = data[NONCE_SIZE+len(decrypted)+secretbox.Overhead:]
	}

	return output
}

func OldDecrypt(password []byte, data []byte) []byte {
	salt, data := data[len(data)-KEY_SIZE:], data[:len(data)-KEY_SIZE]

	key, _, err := DeriveKey(password, salt)
	if err != nil {
		return nil
	}

	var output []byte
	for len(data) > 0 {
		var decryptNonce [NONCE_SIZE]byte
		copy(decryptNonce[:], data[:NONCE_SIZE])

		// Define the size of chunk to read
		chunkSize := NONCE_SIZE + BUFFER_SIZE + secretbox.Overhead
		if chunkSize > len(data)+NONCE_SIZE {
			// Read remainder of file if the chunk size exceeds the available data
			chunkSize = len(data)
		}

		// Decrypt and append contents to output
		decrypted, ok := secretbox.Open(nil, data[NONCE_SIZE:chunkSize], &decryptNonce, &key)
		if !ok {
			panic("decryption error")
		}

		output = append(output, decrypted...)
		data = data[NONCE_SIZE+len(decrypted)+secretbox.Overhead:]
	}

	return output
}
