package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var characters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
var numbers = []rune("1234567890")

func GenRandomString(n int) string {
	randStr := GenRandomArray(n, characters)
	return string(randStr)
}

func GenRandomNumbers(n int) string {
	randNums := GenRandomArray(n, numbers)
	return string(randNums)
}

func GenRandomArray(n int, runes []rune) []rune {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}

	return b
}

func GetEnvVar(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}

	return strings.TrimSpace(value)
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

func PrettyPrintStruct(v any) {
	s, _ := json.MarshalIndent(v, "", "\t")
	fmt.Println(string(s))
}
