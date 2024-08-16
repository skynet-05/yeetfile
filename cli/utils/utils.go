package utils

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"yeetfile/cli/styles"
	"yeetfile/shared"
)

var (
	httpErrorCodeFormat = "[code: %d]"
	separator           = "-"
	r                   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

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

// ParseDownloadString processes a URL such as
// "[http(s)://...]this.example.path#<hex key>"
// into separate usable components: the path to the file (this.example.path),
// and a [32]byte key to use for decrypting the encrypted salt from the server.
func ParseDownloadString(tag string) (string, []byte, error) {
	splitURL := strings.Split(tag, "/")
	splitTag := strings.Split(splitURL[len(splitURL)-1], "#")

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
	return CopyBytesToFile([]byte(contents), to)
}

func CopyBytesToFile(contents []byte, to string) error {
	err := os.WriteFile(to, contents, 0o644)
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

func GenerateTitle(s string) string {
	prefix := "YeetFile CLI: "
	verticalEdge := strings.Repeat("═", len(s)+len(prefix)+2)
	title := styles.BoldStyle.Render(fmt.Sprintf(
		"╔"+verticalEdge+"╗\n"+
			"║ %s%s ║\n"+
			"╚"+verticalEdge+"╝", prefix, s))
	return title
}

// GenerateDescription generates a text box with the provided description
func GenerateDescription(desc string, minLen int) string {
	return GenerateDescriptionSection("", desc, minLen)
}

// GenerateDescriptionSection generates a text box with a title positioned
// above the provided description.
func GenerateDescriptionSection(title, desc string, minLen int) string {
	var out string
	split := strings.Split(desc, "\n")
	maxLen := minLen
	for _, s := range split {
		maxLen = max(maxLen, len(s))
	}

	verticalEdge := strings.Repeat("─", maxLen+2)
	out += styles.BoldStyle.Render("┌"+verticalEdge+"┐") + "\n"
	if len(title) > 0 {
		out += styles.BoldStyle.Render("│ ") + title + strings.Repeat(" ", maxLen-len(title)) + styles.BoldStyle.Render(" │") + "\n"
		out += styles.BoldStyle.Render("│ ") + strings.Repeat("-", maxLen) + styles.BoldStyle.Render(" │") + "\n"
	}

	for _, s := range split {
		out += styles.BoldStyle.Render("│ ") + s + (strings.Repeat(" ", maxLen-len(s))) + styles.BoldStyle.Render(" │") + "\n"
	}
	out += styles.BoldStyle.Render("└" + verticalEdge + "┘")

	return out
}

func GetFilenameFromPath(path string) string {
	fullPath := strings.Split(path, string(os.PathSeparator))
	name := fullPath[len(fullPath)-1]
	return name
}

func GenerateListIdxSpacing(length int) string {
	lenStr := strconv.Itoa(length)
	return strings.Repeat(" ", len(lenStr))
}

func GetListIdxSpacing(spacing string, idx, length int) string {
	idxStr := strconv.Itoa(idx)
	lenStr := strconv.Itoa(length)
	return spacing[0 : len(lenStr)-len(idxStr)+1]
}

func LocalTimeFromUTC(utcTime time.Time) time.Time {
	return utcTime.In(time.Now().Location())
}

func ParseHTTPError(response *http.Response) error {
	body, _ := io.ReadAll(response.Body)
	errCode := fmt.Sprintf(httpErrorCodeFormat, response.StatusCode)
	msg := fmt.Sprintf("server error %s: %s", errCode, body)
	return errors.New(msg)
}
