package utils

import (
	"math/rand"
	"os"
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

func StrArrToStr(arr []string) string {
	return "[\"" + strings.Join(arr, "\",\"") + "\"]"
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
