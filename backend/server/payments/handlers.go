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
	"yeetfile/backend/server/subscriptions"
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
// allows existing subscribers to manage their subscription.
func StripeCustomerPortal(w http.ResponseWriter, req *http.Request, id string) {
	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	link, err := stripe.GetCustomerPortalLink(paymentID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, req, link, http.StatusTemporaryRedirect)
}

// StripeCheckout initiates the process for a user adding to their meter
// using Stripe Checkout
func StripeCheckout(w http.ResponseWriter, req *http.Request, id string) {
	itemType := req.URL.Query().Get("type")
	product, err := subscriptions.GetProductByTag(itemType)
	if err != nil {
		http.Error(w, "Invalid product tag", http.StatusBadRequest)
		return
	}

	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// TODO: Need to decide if auto-renew is worth supporting (probably not)
	//autoRenew := false
	//if len(req.URL.Query().Get("autorenew")) > 0 {
	//	autoRenew = true
	//}

	quantity := 1
	if len(req.URL.Query().Get("quantity")) > 0 {
		quantityInt, err := strconv.Atoi(req.URL.Query().Get("quantity"))
		if err == nil && quantityInt > 0 && quantityInt < 12 {
			quantity = quantityInt
		}
	}

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, req.Host)

	checkoutLink, err := stripe.GenerateCheckoutLink(
		product,
		paymentID,
		quantity,
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
	product, err := subscriptions.GetProductByTag(itemType)
	if err != nil {
		http.Error(w, "Invalid product tag", http.StatusBadRequest)
		return
	}

	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if len(product.BTCPayLink) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutLink := fmt.Sprintf("%s?orderId=%s", product.BTCPayLink, paymentID)
	http.Redirect(w, req, checkoutLink, http.StatusTemporaryRedirect)
}
