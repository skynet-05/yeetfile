package utils

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const characters string = "abcdefghijklmnopqrstuvwxyz1234567890"
const numbers string = "1234567890"

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenRandomString(n int) string {
	randStr := GenRandomArray(n, characters)
	return string(randStr)
}

func GenRandomNumbers(n int) string {
	randNums := GenRandomArray(n, numbers)
	return string(randNums)
}

func GenRandomArray(n int, chars string) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(chars))]
	}

	return b
}

func GetEnvVar(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func StrToDuration(str string) time.Duration {
	unit := string(str[len(str)-1])
	length, _ := strconv.Atoi(str[:len(str)-1])

	if unit == "d" {
		return time.Duration(length) * time.Hour * 24
	} else if unit == "h" {
		return time.Duration(length) * time.Hour
	} else if unit == "m" {
		return time.Duration(length) * time.Minute
	} else if unit == "s" {
		return time.Duration(length) * time.Second
	}

	return 0
}

func GeneratePassphrase() string {
	min := 0
	max := len(EFFWordList)

	var words []string

	i := 0
	randNum := strconv.Itoa(r.Intn(10))
	numInsert := r.Intn(3)
	insertBefore := r.Intn(2) != 0
	for i < 3 {
		idx := r.Intn(max-min+1) + min
		word := EFFWordList[idx]

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

	return strings.Join(words, "-")
}

func GenChecksum(data []byte) ([]byte, string) {
	h := sha1.New()
	h.Write(data)

	checksum := h.Sum(nil)
	return checksum, fmt.Sprintf("%x", checksum)
}

func IsEitherEmpty(a string, b string) bool {
	if (len(a) == 0 && len(b) != 0) || (len(a) != 0 && len(b) == 0) {
		return true
	}

	return false
}

func Contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
