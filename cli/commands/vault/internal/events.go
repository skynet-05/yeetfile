package internal

import "yeetfile/cli/models"

type EventStatus int

const (
	StatusInvalid EventStatus = iota
	StatusOk
	StatusCanceled
)

type Event struct {
	Value  string
	Status EventStatus
	Item   models.VaultItem
	Type   RequestType
}
