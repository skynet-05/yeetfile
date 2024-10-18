package internal

import (
	"yeetfile/cli/crypto"
	"yeetfile/cli/models"
)

type View int

const (
	NullView View = iota
	FilesView
	FilePickerView
	NewPassView
	EditPassView
	ViewPassView
	FileViewerView
	ConfirmationView
	NewFolderView
	RenameView
	ShareView
)

type RequestType int

const (
	InvalidRequest RequestType = iota
	UploadFileRequest
	DeleteFileRequest
	ViewFileRequest
	NewFolderRequest
	NewPassRequest
	EditPassRequest
	RenameRequest
	ShareRequest
	DownloadRequest
)

//
//type ViewCallback struct {
//	Caller      View
//	Event       Event
//	Status      CallbackStatus
//	StringValue string
//}

type ViewRequest struct {
	View      View
	Type      RequestType
	Item      models.VaultItem
	CryptoCtx crypto.CryptoCtx
}
