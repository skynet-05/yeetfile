package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"yeetfile/web/db"
	"yeetfile/web/server/payments/btcpay"
	"yeetfile/web/server/payments/stripe"
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
	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if len(itemType) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutLink, ok := stripe.LinkMapping[itemType]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutParams := fmt.Sprintf("?client_reference_id=%s", paymentID)
	http.Redirect(w, req, checkoutLink+checkoutParams, http.StatusTemporaryRedirect)
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
		log.Printf("Error decoding BTCPay webhook request body")
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
	paymentID, err := db.GetPaymentIDByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if len(itemType) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutLink, ok := btcpay.LinkMapping[itemType]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutParams := fmt.Sprintf("?orderId=%s", paymentID)
	http.Redirect(w, req, checkoutLink+checkoutParams, http.StatusTemporaryRedirect)
}
