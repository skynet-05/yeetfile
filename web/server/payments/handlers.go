package payments

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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

	checkoutLink, ok := stripeLinkMapping[itemType]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Make sure the user is a member if adding upgraded storage
	if itemType != typeSub1Month && itemType != typeSub1Year {
		user, err := db.GetUserByPaymentID(paymentID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if user.MemberExp.Before(time.Now()) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	checkoutParams := fmt.Sprintf("?client_reference_id=%s", paymentID)
	http.Redirect(w, req, checkoutLink+checkoutParams, http.StatusTemporaryRedirect)
}
