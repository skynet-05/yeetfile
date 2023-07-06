package utils

import (
	"math/rand"
	"strings"
	"time"
)

const letters string = "abcdefghijklmnopqrstuvwxyz"

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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
