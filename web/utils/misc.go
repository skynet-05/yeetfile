package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var characters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
var numbers = []rune("1234567890")

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

func GetEnvVar(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}

	return strings.TrimSpace(value)
}

func GetEnvVarInt(key string, fallback int) int {
	value := GetEnvVar(key, "")
	if value == "" {
		return fallback
	}

	num, err := strconv.Atoi(key)
	if err != nil {
		return fallback
	}

	return num
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

// IsEitherEmpty returns true if one string is empty ("") but not the other
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

// GetStructFromFormOrJSON takes a struct and an http request and pulls out
// values from either an http form or a json request body.
func GetStructFromFormOrJSON[T any](t *T, req *http.Request) (T, error) {
	_ = req.ParseForm()
	hasForm := false

	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.Type().NumField(); i++ {
		// Skip fields without json tag
		if tag, ok := val.Type().Field(i).Tag.Lookup("json"); ok {
			formVal := req.FormValue(tag)
			if len(formVal) == 0 {
				fmt.Println("Missing tag: " + tag)
				break
			}

			hasForm = true
			switch val.Field(i).Type().Kind() {
			case reflect.String:
				val.Field(i).SetString(formVal)
				break
			case reflect.Int:
				intVal, _ := strconv.Atoi(formVal)
				val.Field(i).SetInt(int64(intVal))
				break
			case reflect.Bool:
				boolVal, _ := strconv.ParseBool(formVal)
				val.Field(i).SetBool(boolVal)
				break
			case reflect.Float32:
				fallthrough
			case reflect.Float64:
				floatVal, _ := strconv.ParseFloat(formVal, 64)
				val.Field(i).SetFloat(floatVal)
				break
			}
		}
	}

	if !hasForm {
		fmt.Println("Trying to decode")
		err := json.NewDecoder(req.Body).Decode(&t)
		if err != nil {
			return *t, err
		}
	}

	return *t, nil
}
