package signup

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/requests"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

// CreateSignupRequest generates all necessary keys, hashes, etc for initial
// signup. Note that for signup requests without email, an empty signup struct
// is valid since the request has to be generated after the server provides
// an account ID for the user.
func CreateSignupRequest(identifier, password string) shared.Signup {
	if len(identifier) == 0 {
		return shared.Signup{}
	}

	userKey, loginKeyHash := crypto.GenerateUserKeys(identifier, password)
	privateKey, publicKey, err := crypto.GenerateRSAKeyPair()
	utils.HandleCLIError("error generating key pair", err)

	protectedKey, _ := crypto.EncryptChunk(userKey, privateKey)
	rootFolderKey, _ := crypto.GenerateRandomKey()
	protectedRootFolderKey, _ := crypto.EncryptRSA(publicKey, rootFolderKey)

	return shared.Signup{
		Identifier:    identifier,
		LoginKeyHash:  loginKeyHash,
		PublicKey:     publicKey,
		ProtectedKey:  protectedKey,
		RootFolderKey: protectedRootFolderKey,
	}
}

func CreateVerificationRequest(identifier, password, code string) shared.VerifyAccount {
	signup := CreateSignupRequest(identifier, password)
	return shared.VerifyAccount{
		ID:            signup.Identifier,
		Code:          code,
		LoginKeyHash:  signup.LoginKeyHash,
		ProtectedKey:  signup.ProtectedKey,
		PublicKey:     signup.PublicKey,
		RootFolderKey: signup.RootFolderKey,
	}
}

func SubmitSignupForm(signup shared.Signup) (shared.SignupResponse, error) {
	reqData, err := json.Marshal(signup)
	if err != nil {
		return shared.SignupResponse{}, err
	}

	url := endpoints.Signup.Format(config.UserConfig.Server)
	response, err := requests.PostRequest(url, reqData)
	if err != nil {
		return shared.SignupResponse{}, err
	}

	decoder := json.NewDecoder(response.Body)
	var signupResponse shared.SignupResponse
	err = decoder.Decode(&signupResponse)
	if err != nil {
		return shared.SignupResponse{}, err
	} else if len(signupResponse.Error) > 0 {
		return shared.SignupResponse{}, errors.New(signupResponse.Error)
	}

	return signupResponse, nil
}

// VerifyEmailAccount verifies an email user account, which is the final step
// of the email signup process.
func VerifyEmailAccount(email, code string) error {
	url := endpoints.VerifyEmail.Format(config.UserConfig.Server)
	url += fmt.Sprintf("?email=%s&code=%s", email, code)

	response, err := requests.GetRequest(url)
	if err != nil {
		return err
	} else if response.StatusCode >= http.StatusBadRequest {
		if response.StatusCode == http.StatusUnauthorized {
			return errors.New("incorrect verification code")
		}
		return errors.New("server error")
	}

	return nil
}

// FinalizeAccount verifies an ID-only user account and finishes creating it
func FinalizeAccount(account shared.VerifyAccount) error {
	reqData, err := json.Marshal(account)
	if err != nil {
		return err
	}

	url := endpoints.VerifyAccount.Format(config.UserConfig.Server)
	response, err := requests.PostRequest(url, reqData)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusUnauthorized {
			return errors.New("incorrect verification code")
		}
		errStr := fmt.Sprintf("error %d: %v\n", response.StatusCode, response.Body)
		return errors.New(errStr)
	}

	return nil
}
