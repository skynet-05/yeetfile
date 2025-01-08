package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"yeetfile/backend/db"
	"yeetfile/backend/server/payments/btcpay"
	"yeetfile/backend/server/payments/stripe"
	"yeetfile/backend/server/upgrades"
	"yeetfile/backend/utils"
	"yeetfile/shared"
)

// StripeWebhook handles relevant incoming webhook events from Stripe related
// to purchasing storage
func StripeWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the incoming event against the signature header
	signature := req.Header.Get("Stripe-Signature")
	event, err := stripe.ValidateEvent(payload, signature)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process the event received from stripe
	err = stripe.ProcessEvent(event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// StripeCustomerPortal redirects users to the Stripe customer portal, which
// allows existing subscribers to manage their account.
// TODO: Not currently used, should probably be removed in the future
//func StripeCustomerPortal(w http.ResponseWriter, req *http.Request, id string) {
//	paymentID, err := db.GetPaymentIDByUserID(id)
//	if err != nil {
//		w.WriteHeader(http.StatusForbidden)
//		return
//	}
//
//	link, err := stripe.GetCustomerPortalLink(paymentID)
//	if err != nil {
//		w.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	http.Redirect(w, req, link, http.StatusTemporaryRedirect)
//}

// StripeCheckout initiates the process for a user adding to their meter
// using Stripe Checkout
func StripeCheckout(w http.ResponseWriter, req *http.Request, id string) {
	var selectedUpgrades []shared.Upgrade

	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	extractUpgrade := func(paramPrefix string) (shared.Upgrade, error) {
		upgradeParam := fmt.Sprintf("%s-upgrade", paramPrefix)
		quantityParam := fmt.Sprintf("%s-quantity", paramPrefix)

		if req.URL.Query().Has(upgradeParam) {
			upgradeTag := req.URL.Query().Get(upgradeParam)
			upgrade, err := upgrades.GetUpgradeByTag(
				upgradeTag,
				upgrades.GetAllUpgrades())
			if err != nil {
				return shared.Upgrade{}, err
			}

			if req.URL.Query().Has(quantityParam) {
				quantityStr := req.URL.Query().Get(quantityParam)
				quantity, err := strconv.Atoi(quantityStr)
				if err != nil {
					return shared.Upgrade{}, err
				}

				upgrade.Quantity = quantity
			} else {
				upgrade.Quantity = 1
			}

			return upgrade, nil
		}

		return shared.Upgrade{}, nil
	}

	sendUpgrade, err := extractUpgrade("send")
	if err != nil {
		log.Println("Error processing requested send upgrade", err)
		http.Error(w, "Error processing requested send upgrade", http.StatusBadRequest)
		return
	} else if len(sendUpgrade.Tag) > 0 {
		selectedUpgrades = append(selectedUpgrades, sendUpgrade)
	}

	vaultUpgrade, err := extractUpgrade("vault")
	if err != nil {
		log.Println("Error processing requested vault upgrade", err)
		http.Error(w, "Error processing requested vault upgrade", http.StatusBadRequest)
		return
	} else if len(vaultUpgrade.Tag) > 0 {
		selectedUpgrades = append(selectedUpgrades, vaultUpgrade)
	}

	scheme := "http"
	if utils.IsTLSReq(req) {
		scheme = "https"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, req.Host)
	checkoutLink, err := stripe.GenerateCheckoutLink(
		selectedUpgrades,
		paymentID,
		baseURL)

	if err != nil {
		http.Error(w, "Error generating checkout link", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, checkoutLink, http.StatusTemporaryRedirect)
}

// BTCPayWebhook handles relevant incoming webhook events from BTCPay
func BTCPayWebhook(w http.ResponseWriter, req *http.Request) {
	bodyBytes, isValid := btcpay.IsValidRequest(w, req)
	if !isValid {
		log.Printf("Error validating BTCPay webhook event, ignoring")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reader := bytes.NewReader(bodyBytes)
	decoder := json.NewDecoder(reader)
	var settledInvoice btcpay.Invoice
	err := decoder.Decode(&settledInvoice)
	if err != nil {
		log.Printf("Error decoding BTCPay webhook request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = btcpay.FinalizeInvoice(settledInvoice)
	if err != nil {
		log.Printf("Error finalizing BTCPay invoice: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// BTCPayCheckout generates an invoice for the requested product/upgrade
func BTCPayCheckout(w http.ResponseWriter, req *http.Request, id string) {
	itemType := req.URL.Query().Get("type")
	quantity := req.URL.Query().Get("quantity")
	upgrade, err := upgrades.GetUpgradeByTag(itemType, upgrades.GetAllUpgrades())
	if err != nil {
		http.Error(w, "Invalid upgrade tag", http.StatusBadRequest)
		return
	}

	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if len(upgrade.BTCPayLink) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutLink := fmt.Sprintf("%s?orderId=%s&quantity=%s", upgrade.BTCPayLink, paymentID, quantity)
	http.Redirect(w, req, checkoutLink, http.StatusTemporaryRedirect)
}
