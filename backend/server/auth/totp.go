package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"image/jpeg"
	"log"
	"strings"
	"yeetfile/backend/config"
	"yeetfile/backend/crypto"
	"yeetfile/backend/db"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

var AlreadyHasSecretErr = errors.New("user already has a totp secret")
var IncorrectCodeErr = errors.New("incorrect totp code")

func generateUserTotp(userID string) (shared.NewTOTP, error) {
	secret, err := db.GetUserSecret(userID)
	if err != nil {
		return shared.NewTOTP{}, err
	} else if secret != nil && len(secret) > 0 {
		return shared.NewTOTP{}, AlreadyHasSecretErr
	}

	issuer := config.YeetFileConfig.Domain
	if len(issuer) == 0 {
		issuer = "yeetfile (dev)"
	} else if strings.Contains(issuer, "http") {
		issuer = strings.ReplaceAll(issuer, "http://", "")
		issuer = strings.ReplaceAll(issuer, "https://", "")
	}

	accountName, err := db.GetUserPublicName(userID)
	if err != nil {
		return shared.NewTOTP{}, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})

	img, err := key.Image(200, 200)
	if err != nil {
		return shared.NewTOTP{}, err
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 100}); err != nil {
		return shared.NewTOTP{}, err
	}

	return shared.NewTOTP{
		URI:      key.URL(),
		Secret:   key.Secret(),
		B64Image: base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

// validateTOTP checks the provided code against the stored user secret, or
// against one of the recovery code hashes if the provided code length matches
// constants.RecoveryCodeLen.
func validateTOTP(encSecret []byte, code, userID string) error {
	if len(code) == 0 {
		return Missing2FAErr
	} else if len(code) == 6 {
		decSecret, err := crypto.Decrypt(encSecret)
		if err != nil {
			return err
		}

		valid2FA := totp.Validate(code, decSecret)
		if !valid2FA {
			return Failed2FAErr
		}

		return nil
	} else if len(code) == constants.RecoveryCodeLen {
		hashes, err := db.GetUserRecoveryCodeHashes(userID)
		if err != nil {
			return err
		}

		found := -1
		for i, hash := range hashes {
			hashBytes, err := base64.StdEncoding.DecodeString(hash)
			if err != nil {
				return err
			}

			err = bcrypt.CompareHashAndPassword(hashBytes, []byte(code))
			if err == nil {
				found = i
				break
			}
		}

		if found < 0 {
			return Failed2FAErr
		}

		// Remove the found hash
		hashes[found] = hashes[len(hashes)-1]
		hashes = hashes[:len(hashes)-1]
		err = db.SetUserRecoveryCodeHashes(userID, hashes)
		return err
	} else {
		errMsg := fmt.Sprintf("invalid code length %d", len(code))
		return errors.New(errMsg)
	}
}

func removeTOTP(userID, code string) error {
	encSecret, err := db.GetUserSecret(userID)
	if err != nil {
		return err
	}

	err = validateTOTP(encSecret, code, userID)
	if err != nil {
		return err
	}

	err = db.RemoveUser2FA(userID)
	return err
}

func setTOTP(userID string, set shared.SetTOTP) (shared.SetTOTPResponse, error) {
	valid := totp.Validate(set.Code, set.Secret)
	if !valid {
		return shared.SetTOTPResponse{}, IncorrectCodeErr
	}

	var recoveryCodes [6]string
	for i := range recoveryCodes {
		code := shared.GenRandomString(constants.RecoveryCodeLen)
		recoveryCodes[i] = code
	}

	var hashedCodes [6]string
	for i := range hashedCodes {
		byteCode := []byte(recoveryCodes[i])
		hash, err := bcrypt.GenerateFromPassword(byteCode, 8)
		if err != nil {
			return shared.SetTOTPResponse{}, err
		}

		hashedCodes[i] = base64.StdEncoding.EncodeToString(hash)
	}

	encSecret, err := crypto.Encrypt(set.Secret)
	if err != nil {
		return shared.SetTOTPResponse{}, err
	}

	err = db.SetUserSecret(userID, encSecret)
	if err != nil {
		return shared.SetTOTPResponse{}, err
	}

	err = db.SetUserRecoveryCodeHashes(userID, hashedCodes[:])
	if err != nil {
		recoveryErr := db.RemoveUser2FA(userID)
		if recoveryErr != nil {
			log.Printf(
				"Error resetting user 2fa during err: %v\n",
				recoveryErr)
		}
		return shared.SetTOTPResponse{}, err
	}

	return shared.SetTOTPResponse{RecoveryCodes: recoveryCodes}, nil
}
