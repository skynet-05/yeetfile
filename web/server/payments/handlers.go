package payments

import (
	"io"
	"net/http"
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
