package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"yeetfile/shared"
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

// StripeCheckout initiates the process for a user adding to their meter
// using Stripe Checkout
func StripeCheckout(w http.ResponseWriter, req *http.Request, _ string) {
	// Ensure Stripe has already been set up
	if !stripe.Ready {
		log.Println("Stripe checkout requested, but Stripe has not been set up.")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	itemType := req.URL.Query().Get("type")
	paymentID := req.URL.Query().Get("payment_id")
	if len(itemType) == 0 || len(paymentID) == 0 || !db.PaymentIDExists(paymentID) {
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
	bodyBytes, isValid := btcpay.IsValidRequest(req)
	if !isValid {
		log.Printf("Error validating BTCPay webhook event, ignoring")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reader := bytes.NewReader(bodyBytes)
	decoder := json.NewDecoder(reader)
	var settledPayment btcpay.SettledPayment
	err := decoder.Decode(&settledPayment)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = btcpay.FinalizeInvoice(settledPayment)
	if err != nil {
		log.Printf("Error finalizing BTCPay invoice: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// BTCPayCheckout generates an invoice for the requested product/upgrade
func BTCPayCheckout(w http.ResponseWriter, req *http.Request, _ string) {
	// Ensure BTCPay has already been set up
	if !btcpay.Ready {
		log.Println("BTCPay checkout requested, but it has not been set up.")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	itemType := req.URL.Query().Get("type")
	paymentID := req.URL.Query().Get("payment_id")

	isValidID := false
	if len(paymentID) > 0 {
		isValidID = db.PaymentIDExists(paymentID)
	}

	if len(itemType) == 0 || len(paymentID) == 0 || !isValidID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch price of item being purchased
	price, ok := shared.PriceMapping[itemType]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Generate BTCPay invoice w/ checkout link
	invoice, err := btcpay.GenerateBTCPayInvoice(paymentID, price)
	if err != nil {
		fmt.Printf("Error generating BTCPay invoice: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add payment ID to database
	err = db.InsertNewBTCPayOrder(paymentID, invoice.ID, itemType)
	if err != nil {
		fmt.Printf("Error inserting new BTCPay order: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, invoice.CheckoutLink, http.StatusTemporaryRedirect)
}
