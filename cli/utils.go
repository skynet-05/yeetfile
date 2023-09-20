package main

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

// ParseDownloadString processes a URL such as "this.example.path#<hex key>" into
// separate usable components: the path to the file (this.example.path), and
// a [32]byte key to use for decrypting the encrypted salt from the server.
func ParseDownloadString(tag string) (string, []byte, error) {
	splitTag := strings.Split(tag, "#")

	if len(splitTag) != 2 {
		return "", nil, errors.New("invalid download string")
	}

	path := splitTag[0]
	pepper := splitTag[1]

	return path, []byte(pepper), nil
}

func readPassword() []byte {
	pw, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()

	if err != nil {
		fmt.Println("Error reading stdin")
		os.Exit(1)
	}

	return pw
}

// RequestPassword prompts the user for a password
func RequestPassword() []byte {
	fmt.Print("Enter Password: ")
	return readPassword()
}

func ConfirmPassword(pw []byte) bool {
	fmt.Print("Confirm Password: ")
	confirmPw := readPassword()

	return string(confirmPw) == string(pw)
}
