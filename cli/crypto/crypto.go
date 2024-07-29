package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"log"
	"math/big"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

type CryptFunc func([]byte, []byte) ([]byte, error)

// DeriveSendingKey uses PBKDF2 to derive a key for sending a file. Both the
// salt and the pepper can be left nil in order to randomly generate both values.
func DeriveSendingKey(
	password []byte,
	salt []byte,
	pepper []byte,
) ([]byte, []byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, constants.KeySize)
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

// GenerateCLISessionKey uses crypto.rand to generate a random alphanumeric key
// that can be used to secure keys when using the CLI app.
func GenerateCLISessionKey() ([]byte, error) {
	charset := append(shared.Characters, shared.Numbers...)
	result := make([]byte, constants.KeySize)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, err
		}
		result[i] = byte(charset[num.Int64()])
	}

	return result, nil
}

// GenerateRandomKey creates the 256-bit symmetric key used for encrypting files
// that are stored (not sent) in YeetFile. This is always encrypted using the
// master PBKDF2 key before being sent to the server.
func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, constants.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

// DerivePBKDFKey uses PBKDF2 to derive a key from a known password and salt.
func DerivePBKDFKey(password []byte, salt []byte) []byte {
	key := pbkdf2.Key(password, salt, 600000, constants.KeySize, sha256.New)
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
func GenerateLoginKeyHash(userKey []byte, password []byte) []byte {
	loginKey := DerivePBKDFKey(userKey, password)

	h := sha256.New()
	h.Write(loginKey)
	loginKeyHash := h.Sum(nil)

	return loginKeyHash
}

// GenerateRSAKeyPair generates a new RSA-OAEP
func GenerateRSAKeyPair() ([]byte, []byte, error) {
	bitSize := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		fmt.Println("Error generating RSA key:", err)
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fmt.Println("Error encoding private key: ", err)
		return nil, nil, err
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println("Error encoding public key:", err)
		return nil, nil, err
	}

	return privateKeyBytes, publicKeyBytes, nil
}

// EncryptRSA uses a user's public key to encrypt a chunk of data.
func EncryptRSA(key []byte, data []byte) ([]byte, error) {
	hash := sha256.New()
	publicKey, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		log.Println("Error parsing public key: ", err)
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		fmt.Println("Not an RSA public key")
		return nil, err
	}

	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, rsaPublicKey, data, []byte(""))
	if err != nil {
		log.Println("Error encrypting message: ", err)
		return nil, err
	}

	return encrypted, nil
}

// DecryptRSA uses a user's private key to decrypt a chunk of data that's been
// encrypted by their public key.
func DecryptRSA(key []byte, data []byte) ([]byte, error) {
	privateKey, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		log.Printf("Error parsing private key: %v", err)
		return nil, err
	}

	hash := sha256.New()
	decrypted, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		privateKey.(*rsa.PrivateKey),
		data,
		[]byte(""))
	if err != nil {
		log.Printf("Error decrypting message: %v", err)
		return nil, err
	}

	return decrypted, nil
}

// EncryptChunk encrypts a chunk of data using either the sending or storage key.
// Returns the encrypted chunk of data.
func EncryptChunk(key []byte, data []byte) ([]byte, error) {
	var iv [constants.IVSize]byte
	if _, err := io.ReadFull(rand.Reader, iv[:]); err != nil {
		log.Fatalf("Error generating nonce: %v\n", err)
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	result := aesgcm.Seal(nil, iv[:], data, nil)
	var merged []byte
	merged = append(merged, iv[:]...)
	merged = append(merged, result[:]...)

	return merged, nil
}

// DecryptChunk decrypts an encrypted chunk of data using the provided key. If
// the key is unable to decrypt the data, an error is returned, otherwise the
// decrypted data is returned.
func DecryptChunk(key []byte, chunk []byte) ([]byte, error) {
	iv := chunk[:constants.IVSize]
	data := chunk[constants.IVSize:]

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

// GenerateUserKeys generates the main user key as well as the login key hash,
// which is generated from the user key. Returns the user key and login key hash.
func GenerateUserKeys(identifier, password string) ([]byte, []byte) {
	userKey := GenerateUserKey([]byte(identifier), []byte(password))
	loginKeyHash := GenerateLoginKeyHash(userKey, []byte(password))

	return userKey, loginKeyHash
}
