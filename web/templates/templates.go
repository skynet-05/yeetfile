package templates

import "embed"

//go:embed *.html
var HTML embed.FS

type HomePage struct {
	LoggedIn bool
}
