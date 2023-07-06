package utils

import "math/rand"

const letters string = "abcdefghijklmnopqrstuvwxyz"

func GenRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
