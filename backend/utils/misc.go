package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

func Log(msg string) {
	if GetEnvVar("YEETFILE_DEBUG", "0") == "1" {
		log.Println(msg)
	}
}

func Logf(msg string, a ...any) {
	if GetEnvVar("YEETFILE_DEBUG", "0") == "1" {
		log.Printf(msg, a...)
	}
}

func GetEnvVar(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}

	return strings.TrimSpace(value)
}

func GetEnvVarInt(key string, fallback int) int {
	value := GetEnvVar(key, strconv.Itoa(fallback))
	if value == "" {
		return fallback
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return num
}

func GetEnvVarBool(key string, fallback bool) bool {
	value := GetEnvVar(key, "")
	value = strings.ToLower(value)
	if value == "" {
		return fallback
	} else if value == "0" || value == "n" {
		return false
	} else if value == "1" || value == "y" {
		return true
	}

	return fallback
}

func StrToDuration(str string, isDebug bool) time.Duration {
	unit := string(str[len(str)-1])
	length, _ := strconv.Atoi(str[:len(str)-1])

	if unit == "d" {
		return time.Duration(length) * time.Hour * 24
	} else if unit == "h" {
		return time.Duration(length) * time.Hour
	} else if unit == "m" {
		return time.Duration(length) * time.Minute
	} else if unit == "s" {
		if !isDebug {
			// N sec expiry is only available in debug mode
			return time.Minute
		}

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

// IsStructMissingAnyField checks to see if any generic struct is missing a
// values in its string or array fields. Numeric fields are not checked since
// 0 is a valid field value.
func IsStructMissingAnyField(s interface{}) bool {
	val := reflect.ValueOf(s)
	for i := 0; i < val.Type().NumField(); i++ {
		switch val.Field(i).Type().Kind() {
		case reflect.String:
			fallthrough
		case reflect.Slice:
			if val.Field(i).Len() == 0 {
				return true
			}
			break
		}
	}

	return false
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
		// No HTML form, should decode instead
		err := json.NewDecoder(req.Body).Decode(&t)
		if err != nil {
			return *t, err
		}
	}

	return *t, nil
}

func CheckDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func ParseSizeString(str string) int {
	pattern := regexp.MustCompile(`^(\d+)([a-zA-Z]+)$`)
	matches := pattern.FindStringSubmatch(str)

	if len(matches) == 3 {
		numStr := matches[1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			Logf("Error converting number: %v\n", err)
			return 0
		}

		letters := strings.ToUpper(matches[2])

		switch letters[0] {
		case 'T': // Terabyte
			return 1024 * 1024 * 1024 * 1024 * num
		case 'G': // Gigabyte
			return 1024 * 1024 * 1024 * num
		case 'M': // Megabyte
			return 1024 * 1024 * num
		case 'K': // Kilobyte
			return 1024 * num
		default:
			return num
		}
	} else {
		Logf("No match found for size string: %s\n", str)
	}

	return 0
}

func HandleError(w http.ResponseWriter, err error, statusCode int, message string) bool {
	if err != nil {
		Log(fmt.Sprintf("%s: %v\n", message, err))
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(message))
		return true
	}

	return false
}

// LimitedReader reads the request body, limited to max chunk size + encryption
// overhead + 1024 bytes. This is big enough for all requests made to the
// YeetFile API.
func LimitedReader(w http.ResponseWriter, body io.ReadCloser) ([]byte, error) {
	limitedBody := http.MaxBytesReader(
		w,
		body,
		int64(constants.ChunkSize+constants.TotalOverhead+1024))
	return io.ReadAll(limitedBody)
}

func GetTrailingURLSegments(endpoint endpoints.Endpoint, path string) []string {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}

	endpointBase := strings.ReplaceAll(string(endpoint), "/*", "")
	path = strings.Replace(path, endpointBase, "", 1)

	if strings.HasSuffix(path, string(endpoint)) {
		// There is no trailing segment, it ends with the base endpoint
		return []string{}
	}

	path = strings.TrimPrefix(path, "/")
	return strings.Split(path, "/")
}
