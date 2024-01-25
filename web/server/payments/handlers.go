package payments

import (
	"fmt"
	"io"
	"net/http"
	"yeetfile/web/db"
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
	event, err := validateStripeEvent(payload, req.Header.Get("Stripe-Signature"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process the event received from stripe
	err = processStripeEvent(event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// StripeCheckout initiates the process for a user adding to their meter
// using Stripe Checkout
func StripeCheckout(w http.ResponseWriter, req *http.Request) {
	// Ensure Stripe has already been set up
	if !stripeReady {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	size := req.URL.Query().Get("size")
	paymentID := req.URL.Query().Get("payment_id")
	if len(size) == 0 || len(paymentID) == 0 || !db.PaymentIDExists(paymentID) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutLink, ok := stripeLinkMapping[size]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	checkoutParams := fmt.Sprintf("?client_reference_id=%s", paymentID)
	http.Redirect(w, req, checkoutLink+checkoutParams, http.StatusTemporaryRedirect)
}
