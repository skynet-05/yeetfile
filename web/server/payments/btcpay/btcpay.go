package btcpay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"yeetfile/web/server/payments"
)

var apiKey = os.Getenv("YEETFILE_BTCPAY_API_KEY")
var storeID = os.Getenv("YEETFILE_BTCPAY_STORE_ID")
var serverURL = os.Getenv("YEETFILE_BTCPAY_SERVER_URL")

var Ready = true

// FIXME: Test values, needs updating for prod
var btcPayPriceMapping = map[string]string{
	payments.TypeSub1Month: "0.01",
	payments.TypeSub1Year:  "0.01",
	payments.Type100GB:     "0.01",
	payments.Type500GB:     "0.01",
	payments.Type1TB:       "0.01",
}

// IsValidRequest validates incoming webhook events from BTCPay Server
func IsValidRequest(req *http.Request) bool {
	secret := os.Getenv("YEETFILE_BTCPAY_WEBHOOK_SECRET")
	sig := req.Header.Get("BTCPAY-SIG")
	if len(sig) == 0 || len(secret) == 0 {
		return false
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading BTCPay webhook body")
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(reqBody)
	expectedMAC := fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
	return sig == expectedMAC
}

// sendRequest sends a request to BTCPay Server with the correct authentication
// headers set up.
func sendRequest(method string, path string, data []byte) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s/%s", serverURL, path)
	req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", apiKey))

	resp, err := new(http.Transport).RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// init ensures that the necessary values for interacting with BTCPay Server
// have already been defined.
func init() {
	if len(apiKey) == 0 || len(storeID) == 0 || len(serverURL) == 0 {
		Ready = false
	}
}
