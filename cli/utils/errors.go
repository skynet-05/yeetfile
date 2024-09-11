package utils

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"os"
	"yeetfile/cli/styles"
)

func HandleCLIError(msg string, err error) {
	if err == nil {
		return
	} else if err == huh.ErrUserAborted {
		os.Exit(0)
	}

	styles.PrintErrStr(fmt.Sprintf("ERROR: %s - %v\n", msg, err))
	os.Exit(1)
}
