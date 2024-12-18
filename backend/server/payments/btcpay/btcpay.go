package btcpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"yeetfile/backend/config"
	"yeetfile/backend/utils"
)

// IsValidRequest validates incoming webhook events from BTCPay Server
func IsValidRequest(w http.ResponseWriter, req *http.Request) ([]byte, bool) {
	secret := config.YeetFileConfig.BTCPayBilling.WebhookSecret
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
