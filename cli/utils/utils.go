package utils

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// StringPrompt prompts the user for string input
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

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

// ConfirmPassword prompts the user for a password again, but checks against
// the provided password bytes to confirm that they're the same.
func ConfirmPassword(pw []byte) bool {
	fmt.Print("Confirm Password: ")
	confirmPw := readPassword()

	return string(confirmPw) == string(pw)
}

func CopyToFile(contents string, to string) error {
	err := os.WriteFile(to, []byte(contents), 0644)
	if err != nil {
		return err
	}

	return err
}

func StrFlag(strVar *string, name string, fallback string, args []string) {
	if len(*strVar) > 0 {
		// This var has already been set
		return
	}

	flagNameA := fmt.Sprintf("-%s", string(name[0]))
	flagNameB := fmt.Sprintf("--%s", name)

	for idx, arg := range args {
		if arg == flagNameA || arg == flagNameB {
			if idx > len(args)-1 {
				// Invalid flag value
				return
			}
			*strVar = args[idx+1]
			return
		}
	}

	*strVar = fallback
}

func BoolFlag(boolVar *bool, name string, fallback bool, args []string) {
	if *boolVar {
		// This var has already been set
		return
	}

	flagNameA := fmt.Sprintf("-%s", string(name[0]))
	flagNameB := fmt.Sprintf("--%s", name)

	for _, arg := range args {
		if arg == flagNameA || arg == flagNameB {
			*boolVar = true
			return
		}
	}

	*boolVar = fallback
}

func IntFlag(intVar *int, name string, fallback int, args []string) {
	if *intVar != 0 {
		// This var has already been set
		return
	}

	flagNameA := fmt.Sprintf("-%s", string(name[0]))
	flagNameB := fmt.Sprintf("--%s", name)

	for idx, arg := range args {
		if arg == flagNameA || arg == flagNameB {
			if idx > len(args)-1 {
				// Invalid flag value
				return
			}

			*intVar, _ = strconv.Atoi(args[idx+1])
			return
		}
	}

	*intVar = fallback
}
