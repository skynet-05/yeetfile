package btcpay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"yeetfile/web/config"
	"yeetfile/web/server/subscriptions"
	"yeetfile/web/utils"
)

var apiKey = os.Getenv("YEETFILE_BTCPAY_API_KEY")
var storeID = os.Getenv("YEETFILE_BTCPAY_STORE_ID")
var serverURL = os.Getenv("YEETFILE_BTCPAY_SERVER_URL")

var LinkMapping = map[string]string{
	subscriptions.MonthlyNovice: config.YeetFileConfig.BTCPayBilling.
		SubNoviceMonthlyLink,
	subscriptions.MonthlyRegular: config.YeetFileConfig.BTCPayBilling.
		SubRegularMonthlyLink,
	subscriptions.MonthlyAdvanced: config.YeetFileConfig.BTCPayBilling.
		SubAdvancedMonthlyLink,

	subscriptions.YearlyNovice: config.YeetFileConfig.BTCPayBilling.
		SubNoviceYearlyLink,
	subscriptions.YearlyRegular: config.YeetFileConfig.BTCPayBilling.
		SubRegularYearlyLink,
	subscriptions.YearlyAdvanced: config.YeetFileConfig.BTCPayBilling.
		SubAdvancedYearlyLink,
}

// IsValidRequest validates incoming webhook events from BTCPay Server
func IsValidRequest(w http.ResponseWriter, req *http.Request) ([]byte, bool) {
	secret := os.Getenv("YEETFILE_BTCPAY_WEBHOOK_SECRET")
	sig := req.Header.Get("BTCPAY-SIG")
	if len(sig) == 0 || len(secret) == 0 {
		return nil, false
	}

	reqBody, err := utils.LimitedReader(w, req.Body)
	if err != nil {
		log.Printf("Error reading BTCPay webhook body")
		return nil, false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(reqBody)
	expectedMAC := fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil)))
	return reqBody, sig == expectedMAC
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
