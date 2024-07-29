package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"yeetfile/cli/config"
	"yeetfile/cli/crypto"
	"yeetfile/cli/requests"
	"yeetfile/cli/utils"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

// LogIn logs into YeetFile by using the provided identifier and password to
// generate the login key hash, and stores the user's key pair in their config
// directory
func LogIn(identifier, password string, vaultKey []byte) error {
	userKey, loginKeyHash := crypto.GenerateUserKeys(identifier, password)

	login := shared.Login{
		Identifier:   identifier,
		LoginKeyHash: loginKeyHash,
	}

	reqData, err := json.Marshal(login)
	utils.HandleCLIError("failed to marshal login request", err)

	url := endpoints.Login.Format(config.UserConfig.Server)
	response, err := requests.PostRequest(url, reqData)
	utils.HandleCLIError("failed to send login request to server", err)

	body, err := io.ReadAll(response.Body)
	utils.HandleCLIError("failed to read login response body", err)

	if response.StatusCode != http.StatusOK {
		// Unlike other errors, we want this one returned back to the
		// Login model, since it generally indicates a problem with
		// their credentials, not with the CLI or server.
		errMsg := fmt.Sprintf("Error %d: %s\n", response.StatusCode, body)
		return errors.New(errMsg)
	}

	var loginResponse shared.LoginResponse
	err = json.Unmarshal(body, &loginResponse)
	utils.HandleCLIError("failed to unmarshal login response", err)

	privateKey, err := crypto.DecryptChunk(userKey, loginResponse.ProtectedKey)
	utils.HandleCLIError("failed to decrypt private key", err)

	encPrivateKey, _ := crypto.EncryptChunk(vaultKey, privateKey)
	err = config.UserConfigPaths.SetKeys(encPrivateKey, loginResponse.PublicKey)
	utils.HandleCLIError("failed to write encrypted private key and public key to config dir", err)

	cookies := response.Cookies()
	if len(cookies) > 0 {
		err = config.UserConfigPaths.SetSession(cookies[0].Value)
		if err != nil {
			utils.HandleCLIError("error initializing new session", err)
		}
	}

	return nil
}
