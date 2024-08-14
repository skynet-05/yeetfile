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
func CreateSignupRequest(identifier, password, serverPw string) shared.Signup {
	if len(identifier) == 0 {
		return shared.Signup{}
	}

	signupKeys, err := crypto.GenerateSignupKeys(identifier, password)
	if err != nil {
		utils.HandleCLIError("error generating signup keys", err)
	}

	return shared.Signup{
		Identifier:     identifier,
		LoginKeyHash:   signupKeys.LoginKeyHash,
		PublicKey:      signupKeys.PublicKey,
		ProtectedKey:   signupKeys.ProtectedPrivateKey,
		RootFolderKey:  signupKeys.ProtectedRootFolderKey,
		ServerPassword: serverPw,
	}
}

func CreateVerificationRequest(identifier, password, code string) shared.VerifyAccount {
	signup := CreateSignupRequest(identifier, password, "")
	return shared.VerifyAccount{
		ID:            signup.Identifier,
		Code:          code,
		LoginKeyHash:  signup.LoginKeyHash,
		ProtectedKey:  signup.ProtectedKey,
		PublicKey:     signup.PublicKey,
		RootFolderKey: signup.RootFolderKey,
	}
}
