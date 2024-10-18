package signup

import (
	"yeetfile/cli/crypto"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

// CreateSignupRequest generates all necessary keys, hashes, etc. for initial
// signup. Note that for signup requests without email, an empty signup struct
// is valid since the request has to be generated after the server provides
// an account ID for the user.
func CreateSignupRequest(identifier, password, hint, serverPw string) shared.Signup {
	if len(identifier) == 0 {
		return shared.Signup{}
	}

	signupKeys, err := crypto.GenerateSignupKeys(identifier, password)
	if err != nil {
		utils.HandleCLIError("error generating signup keys", err)
	}

	return shared.Signup{
		Identifier:              identifier,
		LoginKeyHash:            signupKeys.LoginKeyHash,
		PublicKey:               signupKeys.PublicKey,
		ProtectedPrivateKey:     signupKeys.ProtectedPrivateKey,
		ProtectedVaultFolderKey: signupKeys.ProtectedRootFolderKey,
		ServerPassword:          serverPw,
		PasswordHint:            hint,
	}
}

func CreateVerificationRequest(identifier, password, code string) shared.VerifyAccount {
	signup := CreateSignupRequest(identifier, password, "", "")
	return shared.VerifyAccount{
		ID:                      signup.Identifier,
		Code:                    code,
		LoginKeyHash:            signup.LoginKeyHash,
		PublicKey:               signup.PublicKey,
		ProtectedPrivateKey:     signup.ProtectedPrivateKey,
		ProtectedVaultFolderKey: signup.ProtectedVaultFolderKey,
	}
}
