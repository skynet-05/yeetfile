package endpoints

import (
	"fmt"
	"strings"
)

const apiVersion = "v1"

type Endpoint string

var (
	Signup        = genEndpoint("/api/%s/signup")
	Login         = genEndpoint("/api/%s/login")
	Logout        = genEndpoint("/api/%s/logout")
	Account       = genEndpoint("/api/%s/account")
	Forgot        = genEndpoint("/api/%s/forgot")
	Reset         = genEndpoint("/api/%s/reset")
	Session       = genEndpoint("/api/%s/session")
	VerifyAccount = genEndpoint("/api/%s/verify")
	VerifyEmail   = genEndpoint("/verify-email")

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

	ShareFile   = genEndpoint("/api/%s/share/file/*")
	ShareFolder = genEndpoint("/api/%s/share/folder/*")
	PubKey      = genEndpoint("/api/%s/pubkey")
)

var JSVarNameMap = map[Endpoint]string{
	Signup:        "Signup",
	Login:         "Login",
	Logout:        "Logout",
	Forgot:        "Forgot",
	Reset:         "Reset",
	Session:       "Session",
	VerifyAccount: "VerifyAccount",
	VerifyEmail:   "VerifyEmail",

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

	ShareFile:   "ShareFile",
	ShareFolder: "ShareFolder",
	PubKey:      "PubKey",
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
