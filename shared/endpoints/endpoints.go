package endpoints

import (
	"fmt"
	"strings"
)

const apiVersion = "v1"

type Endpoint string

type HTMLEndpoints struct {
	Account        string
	Send           string
	Vault          string
	Login          string
	Signup         string
	Forgot         string
	ChangePassword string
	VerifyEmail    string
}

var HTMLPageEndpoints HTMLEndpoints

var (
	Signup         = genEndpoint("/api/%s/signup")
	Login          = genEndpoint("/api/%s/login")
	Logout         = genEndpoint("/api/%s/logout")
	Account        = genEndpoint("/api/%s/account")
	Forgot         = genEndpoint("/api/%s/forgot")
	Reset          = genEndpoint("/api/%s/reset")
	Session        = genEndpoint("/api/%s/session")
	VerifyAccount  = genEndpoint("/api/%s/verify")
	VerifyEmail    = genEndpoint("/api/%s/verify_email")
	ChangePassword = genEndpoint("/api/%s/change_password")

	VaultRoot   = genEndpoint("/api/%s/vault")
	VaultFolder = genEndpoint("/api/%s/vault/folder/*")
	VaultFile   = genEndpoint("/api/%s/vault/file/*")

	UploadVaultFileMetadata   = genEndpoint("/api/%s/vault/u")
	UploadVaultFileData       = genEndpoint("/api/%s/vault/u/*/*")
	DownloadVaultFileMetadata = genEndpoint("/api/%s/vault/d/*")
	DownloadVaultFileData     = genEndpoint("/api/%s/vault/d/*/*")

	UploadSendFileMetadata   = genEndpoint("/api/%s/send/u")
	UploadSendFileData       = genEndpoint("/api/%s/send/u/*/*")
	UploadSendText           = genEndpoint("/api/%s/send/plaintext")
	DownloadSendFileMetadata = genEndpoint("/api/%s/send/d/*")
	DownloadSendFileData     = genEndpoint("/api/%s/send/d/*/*")

	ShareFile    = genEndpoint("/api/%s/share/file/*")
	ShareFolder  = genEndpoint("/api/%s/share/folder/*")
	PubKey       = genEndpoint("/api/%s/pubkey")
	ProtectedKey = genEndpoint("/api/%s/protectedkey")

	HTMLAccount        = genEndpoint("/account")
	HTMLSend           = genEndpoint("/send")
	HTMLSendDownload   = genEndpoint("/send/*")
	HTMLVault          = genEndpoint("/vault")
	HTMLVaultFolder    = genEndpoint("/vault/*")
	HTMLLogin          = genEndpoint("/login")
	HTMLSignup         = genEndpoint("/signup")
	HTMLForgot         = genEndpoint("/forgot")
	HTMLChangePassword = genEndpoint("/change_password")
	HTMLVerifyEmail    = genEndpoint("/verify_email")
)

var JSVarNameMap = map[Endpoint]string{
	Signup:         "Signup",
	Login:          "Login",
	Logout:         "Logout",
	Forgot:         "Forgot",
	Reset:          "Reset",
	Session:        "Session",
	Account:        "Account",
	VerifyAccount:  "VerifyAccount",
	VerifyEmail:    "VerifyEmail",
	ChangePassword: "ChangePassword",

	VaultRoot:   "VaultRoot",
	VaultFolder: "VaultFolder",
	VaultFile:   "VaultFile",

	UploadVaultFileMetadata:   "UploadVaultFileMetadata",
	UploadVaultFileData:       "UploadVaultFileData",
	DownloadVaultFileMetadata: "DownloadVaultFileMetadata",
	DownloadVaultFileData:     "DownloadVaultFileData",

	UploadSendFileMetadata:   "UploadSendFileMetadata",
	UploadSendFileData:       "UploadSendFileData",
	UploadSendText:           "UploadSendText",
	DownloadSendFileMetadata: "DownloadSendFileMetadata",
	DownloadSendFileData:     "DownloadSendFileData",

	ShareFile:    "ShareFile",
	ShareFolder:  "ShareFolder",
	PubKey:       "PubKey",
	ProtectedKey: "ProtectedKey",

	HTMLAccount:        "HTMLAccount",
	HTMLSend:           "HTMLSend",
	HTMLSendDownload:   "HTMLSendDownload",
	HTMLVault:          "HTMLVault",
	HTMLVaultFolder:    "HTMLVaultFolder",
	HTMLLogin:          "HTMLLogin",
	HTMLSignup:         "HTMLSignup",
	HTMLChangePassword: "HTMLChangePassword",
	HTMLVerifyEmail:    "HTMLVerifyEmail",
}

func genEndpoint(fmtStr string) Endpoint {
	if !strings.Contains(fmtStr, "%s") {
		return Endpoint(fmtStr)
	}

	return Endpoint(fmt.Sprintf(fmtStr, apiVersion))
}

func (e Endpoint) Format(server string, args ...string) string {
	strEndpoint := string(e)
	for _, arg := range args {
		strEndpoint = strings.Replace(strEndpoint, "*", arg, 1)
	}

	// Remove remaining wildcards
	strEndpoint = strings.ReplaceAll(strEndpoint, "*", "")

	server = strings.TrimSuffix(server, "/")
	strEndpoint = strings.TrimPrefix(strEndpoint, "/")
	url := fmt.Sprintf("%s/%s", server, strEndpoint)
	return url
}

func init() {
	HTMLPageEndpoints = HTMLEndpoints{
		Account:        string(HTMLAccount),
		Send:           string(HTMLSend),
		Vault:          string(HTMLVault),
		Login:          string(HTMLLogin),
		Signup:         string(HTMLSignup),
		Forgot:         string(HTMLForgot),
		ChangePassword: string(HTMLChangePassword),
		VerifyEmail:    string(HTMLVerifyEmail),
	}
}
