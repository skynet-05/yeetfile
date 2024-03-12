package shared

import (
	"bufio"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var characters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
var numbers = []rune("1234567890")

func ReadableFileSize(b int) string {
	const unit = 1024
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

// IsPlaintext takes a string determines if it contains non-ascii characters
func IsPlaintext(text string) bool {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		for _, r := range scanner.Text() {
			if r > 127 {
				return false // Non-ASCII character found
			}
		}
	}

	return true
}

// AddDate is similar to time.AddDate, but defaults to maxing out an available
// month's number of days rather than rolling over into the following month.
func AddDate(years int, months int) time.Time {
	now := time.Now()
	future := now.AddDate(years, months, 0)
	if d := future.Day(); d != now.Day() {
		return future.AddDate(0, 0, -d)
	}

	return future
}

func GenRandomStringWithPrefix(n int, prefix string) string {
	randStr := GenRandomArray(n, characters)

	if len(prefix) == 0 {
		return string(randStr)
	}

	return fmt.Sprintf("%s_%s", prefix, string(randStr))
}

func GenRandomString(n int) string {
	randStr := GenRandomArray(n, characters)
	return string(randStr)
}

func GenRandomNumbers(n int) string {
	randNums := GenRandomArray(n, numbers)
	return string(randNums)
}

func GenRandomArray(n int, runes []rune) []rune {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	b := make([]rune, n)
	for i := range b {
		b[i] = runes[r.Intn(len(runes))]
	}

	return b
}
