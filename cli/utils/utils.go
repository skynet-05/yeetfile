package utils

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/term"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"yeetfile/shared"
)

var LineDecorator = "========================================"
var separator = "-"
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

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

// GeneratePassphrase generates a 3 word passphrase with a randomly placed
// number, and each word separated by the `separator` character
func GeneratePassphrase() string {
	min := 0
	max := len(shared.EFFWordList)

	var words []string

	i := 0
	randNum := strconv.Itoa(r.Intn(10))
	numInsert := r.Intn(3)
	insertBefore := r.Intn(2) != 0
	for i < 3 {
		idx := r.Intn(max-min) + min
		word := shared.EFFWordList[idx]

		shouldInsertNum := numInsert == i

		if shouldInsertNum {
			if insertBefore {
				word = randNum + word
			} else {
				word = word + randNum
			}
		}

		words = append(words, word)
		i++
	}

	return strings.Join(words, separator)
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

func ReadableFileSize(b int) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGT"[exp])
}
