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

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
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

func GenFilePath() string {
	min := 0
	max := len(EFFLongWordList)

	var words []string

	i := 0
	for i < 3 {
		idx := r.Intn(max-min+1) + min
		words = append(words, EFFLongWordList[idx])
		i++
	}

	return strings.Join(words, ".")
}

func GenChecksum(data []byte) ([]byte, string) {
	h := sha1.New()
	h.Write(data)

	checksum := h.Sum(nil)
	return checksum, fmt.Sprintf("%x", checksum)
}
