package pass

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"math/rand"
	"strconv"
	"strings"
	"yeetfile/cli/crypto"
	"yeetfile/cli/globals"
	"yeetfile/cli/utils"
	"yeetfile/shared"
)

const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const lower = "abcdefghijklmnopqrstuvwxyz"
const numbers = "1234567890"
const defaultSymbols = "!@#$%^&*"

type passphraseOpts struct {
	wordlist   []string
	shortWords bool
	numWords   int
	separator  string
	capitalize bool
	useNumber  bool
}

var defaultPassphraseOpts = passphraseOpts{
	wordlist:   globals.LongWordlist,
	shortWords: false,
	numWords:   3,
	separator:  ".",
	capitalize: true,
	useNumber:  true,
}

type passwordOpts struct {
	size       int
	useUpper   bool
	useLower   bool
	useNumbers bool
	useSymbols bool
	symbols    string
}

var defaultPasswordOpts = passwordOpts{
	size:       12,
	useUpper:   true,
	useLower:   true,
	useNumbers: true,
	useSymbols: true,
	symbols:    defaultSymbols,
}

func (opts passwordOpts) generateDescription() string {
	return fmt.Sprintf("# characters: %d\n"+
		"A-Z: %v\n"+
		"a-z: %v\n"+
		"0-9: %v\n"+
		"Symbols: %v\n"+
		"Use symbols: %s\n",
		opts.size,
		opts.useUpper,
		opts.useLower,
		opts.useNumbers,
		opts.useSymbols,
		shared.EscapeString(opts.symbols))
}

func (opts passphraseOpts) generateDescription() string {
	return fmt.Sprintf("# words: %d\n"+
		"Short words: %v\n"+
		"Capitalize: %v\n"+
		"Include number: %v\n"+
		"Separator: %s\n",
		opts.numWords,
		opts.shortWords,
		opts.capitalize,
		opts.useNumber,
		shared.EscapeString(opts.separator))
}

// generatePassword generates a password consisting of passwordOpts.size
// characters containing:
//   - A-Z (if passwordOpts.useUpper)
//   - a-z (if passwordOpts.useLower)
//   - 0-9 (if passwordOpts.useNumbers)
//   - passwordOpts.symbols (if passwordOpts.useSymbols)
//
// The function ensures that at least one character from each category is
// included by prepending a random character from each required category to the
// beginning of the string before filling the remainder of the string with
// random characters from all allowed categories. The string is then shuffled to
// make the output less predictable.
func generatePassword(opts passwordOpts) (string, error) {
	if opts.useSymbols && len(opts.symbols) == 0 {
		opts.symbols = defaultSymbols
	}

	getRandomChar := func(chars string) string {
		randNum, _ := crypto.GenerateRandomNumber(len(chars))
		return string(chars[randNum])
	}

	var chars string
	var result []string

	if opts.useUpper {
		chars += upper
		result = append(result, getRandomChar(upper))
	}

	if opts.useLower {
		chars += lower
		result = append(result, getRandomChar(lower))
	}

	if opts.useNumbers {
		chars += numbers
		result = append(result, getRandomChar(numbers))
	}

	if opts.useSymbols {
		chars += opts.symbols
		result = append(result, getRandomChar(opts.symbols))
	}

	i := len(result)
	for i < opts.size {
		result = append(result, getRandomChar(chars))
		i++
	}

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return strings.Join(result, ""), nil
}

// generatePassphrase generates a random passphrase consisting of
// passphraseOpts.numWords words. Each word has a separator string defined by
// passphraseOpts.separator, and can have a random number placed before or after
// any word.
func generatePassphrase(opts passphraseOpts) (string, error) {
	var passphrase string
	var err error
	numIdx := -1

	if opts.useNumber {
		// numIdx establishes where the number should go in the
		// passphrase. Odd indices place the number before a word, and
		// even indices place the number after a word. For example:
		//
		// randomNumber = 8
		// numIdx = 3
		// passphrase words = hideous.monstrosity.abominable
		//                    0     1 2         3 4        5
		//                                      ^
		// output = hideous.monstrosity8.abominable
		numIdx, err = crypto.GenerateRandomNumber((opts.numWords * 2) - 1)
		if err != nil {
			return "", err
		}
	}

	randNum, _ := crypto.GenerateRandomNumber(9)
	randNumStr := strconv.Itoa(randNum)

	i := 0
	for i < opts.numWords {
		if numIdx == i*2 {
			passphrase += randNumStr
		}

		wordIdx, _ := crypto.GenerateRandomNumber(len(opts.wordlist) - 1)
		word := opts.wordlist[wordIdx]

		if opts.capitalize {
			word = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}

		passphrase += word

		if numIdx == (i*2)+1 {
			passphrase += randNumStr
		}

		if i < opts.numWords-1 {
			passphrase += opts.separator
		}

		i++
	}

	return passphrase, nil
}

func generatePassEntryDescription(
	name,
	username string,
	urls []*string,
	password string,
	mode huh.EchoMode,
	notes string,
) string {
	var desc string
	if len(username) == 0 {
		desc = "Username: None\n"
	} else {
		desc = fmt.Sprintf("Username: %s\n", shared.EscapeString(username))
	}

	if len(urls) == 0 {
		desc += "URLs: None\n"
	} else {
		desc += "URLs:\n"
		for _, url := range urls {
			desc += fmt.Sprintf("   - %s\n", shared.EscapeString(*url))
		}
	}

	if len(password) == 0 {
		desc += "Password: None\n"
	} else {
		if mode == huh.EchoModePassword {
			pwStr := strings.Repeat("*", 8)
			desc += fmt.Sprintf("Password: %s\n", shared.EscapeString(pwStr))
		} else {
			desc += fmt.Sprintf("Password: %s\n", shared.EscapeString(password))
		}
	}

	if len(notes) == 0 {
		desc += "Notes: None\n"
	} else {
		desc += fmt.Sprintf("\nNotes:\n%s\n", shared.EscapeString(utils.GenerateWrappedText(notes)))
	}

	return utils.GenerateDescriptionSection(shared.EscapeString(name), desc, 33)
}
