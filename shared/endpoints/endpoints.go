package endpoints

import (
	"fmt"
	"strings"
)

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

type BillingEndpoints struct {
	BTCPayCheckout string
	StripeCheckout string
	StripeManage   string
}

var HTMLPageEndpoints HTMLEndpoints
var BillingPageEndpoints BillingEndpoints

var (
	Signup         = Endpoint("/api/signup")
	Login          = Endpoint("/api/login")
	Logout         = Endpoint("/api/logout")
	Account        = Endpoint("/api/account")
	Forgot         = Endpoint("/api/forgot")
	Reset          = Endpoint("/api/reset")
	Session        = Endpoint("/api/session")
	VerifyAccount  = Endpoint("/api/verify")
	VerifyEmail    = Endpoint("/api/verify_email")
	ChangePassword = Endpoint("/api/change_password")

	VaultRoot   = Endpoint("/api/vault")
	VaultFolder = Endpoint("/api/vault/folder/*")
	VaultFile   = Endpoint("/api/vault/file/*")

	UploadVaultFileMetadata   = Endpoint("/api/vault/u")
	UploadVaultFileData       = Endpoint("/api/vault/u/*/*")
	DownloadVaultFileMetadata = Endpoint("/api/vault/d/*")
	DownloadVaultFileData     = Endpoint("/api/vault/d/*/*")

	UploadSendFileMetadata   = Endpoint("/api/send/u")
	UploadSendFileData       = Endpoint("/api/send/u/*/*")
	UploadSendText           = Endpoint("/api/send/plaintext")
	DownloadSendFileMetadata = Endpoint("/api/send/d/*")
	DownloadSendFileData     = Endpoint("/api/send/d/*/*")

	ShareFile    = Endpoint("/api/share/file/*")
	ShareFolder  = Endpoint("/api/share/folder/*")
	PubKey       = Endpoint("/api/pubkey")
	ProtectedKey = Endpoint("/api/protectedkey")

	StripeWebhook  = Endpoint("/stripe/webhook")
	StripeManage   = Endpoint("/stripe/manage")
	StripeCheckout = Endpoint("/stripe/checkout")
	BTCPayWebhook  = Endpoint("/btcpay/webhook")
	BTCPayCheckout = Endpoint("/btcpay/checkout")

	HTMLAccount        = Endpoint("/account")
	HTMLHome           = Endpoint("/")
	HTMLSend           = Endpoint("/send")
	HTMLSendDownload   = Endpoint("/send/*")
	HTMLVault          = Endpoint("/vault")
	HTMLVaultFolder    = Endpoint("/vault/*")
	HTMLLogin          = Endpoint("/login")
	HTMLSignup         = Endpoint("/signup")
	HTMLForgot         = Endpoint("/forgot")
	HTMLChangePassword = Endpoint("/change_password")
	HTMLVerifyEmail    = Endpoint("/verify_email")
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

	HTMLHome:           "HTMLHome",
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

	BillingPageEndpoints = BillingEndpoints{
		BTCPayCheckout: string(BTCPayCheckout),
		StripeCheckout: string(StripeCheckout),
		StripeManage:   string(StripeManage),
	}
}
