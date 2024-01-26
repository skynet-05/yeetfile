package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
	"os"
	"time"
	"yeetfile/web/db"
	"yeetfile/web/mail"
	"yeetfile/web/utils"
)

var stripeReady = true

var stripeSubMonthID = os.Getenv("YEETFILE_STRIPE_SUB_MONTH_ID")
var stripeSubYearID = os.Getenv("YEETFILE_STRIPE_SUB_YEAR_ID")
var stripe100GBID = os.Getenv("YEETFILE_STRIPE_100GB_ID")
var stripe500GBID = os.Getenv("YEETFILE_STRIPE_500GB_ID")
var stripe1TBID = os.Getenv("YEETFILE_STRIPE_1TB_ID")
var stripeSubMonthLink = os.Getenv("YEETFILE_STRIPE_SUB_MONTH_LINK")
var stripeSubYearLink = os.Getenv("YEETFILE_STRIPE_SUB_YEAR_LINK")
var stripe100GBLink = os.Getenv("YEETFILE_STRIPE_100GB_LINK")
var stripe500GBLink = os.Getenv("YEETFILE_STRIPE_500GB_LINK")
var stripe1TBLink = os.Getenv("YEETFILE_STRIPE_1TB_LINK")

var stripeRequirements = []string{
	stripe100GBID, stripe500GBID, stripe1TBID, stripeSubMonthID, stripeSubYearID,
	stripe100GBLink, stripe500GBLink, stripe1TBLink, stripeSubMonthLink, stripeSubYearLink,
}

var stripeLinkMapping = map[string]string{
	typeSub1Month: stripeSubMonthLink,
	typeSub1Year:  stripeSubYearLink,
	type100gb:     stripe100GBLink,
	type500gb:     stripe500GBLink,
	type1tb:       stripe1TBLink,
}

var stripeDescMap = map[string]string{
	stripeSubMonthID: "1 Month YeetFile Membership",
	stripeSubYearID:  "1 Year YeetFile Membership",
	stripe100GBID:    "YeetFile 100GB Transfer Upgrade",
	stripe500GBID:    "YeetFile 500GB Transfer Upgrade",
	stripe1TBID:      "YeetFile 1TB Transfer Upgrade",
}

// stripeProductAmounts maps product IDs to their respective amounts of storage
// that they grant a user
var stripeProductAmounts = map[string]int{
	os.Getenv("YEETFILE_STRIPE_100GB_ID"): 107_374_182_400,   // 100GB
	os.Getenv("YEETFILE_STRIPE_500GB_ID"): 536_870_912_000,   // 500GB
	os.Getenv("YEETFILE_STRIPE_1TB_ID"):   1_073_741_824_000, // 1TB
}

// processStripeEvent receives an input stripe.Event and determines if/how a
// user's meter should be updated depending on the product they purchased.
func processStripeEvent(event stripe.Event) error {
	// Currently only successful checkouts are handled by the webhook
	if event.Type != "checkout.session.completed" {
		return nil
	}

	var checkoutSession stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &checkoutSession)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
		return err
	}

	// Fetch line items to figure out product IDs and quantities of
	// each product purchased
	params := &stripe.CheckoutSessionListLineItemsParams{}
	params.Session = stripe.String(checkoutSession.ID)
	lineItems := session.ListLineItems(params)

	for lineItems.Next() {
		utils.PrettyPrintStruct(checkoutSession)
		refID := checkoutSession.ClientReferenceID
		var intentID string
		if checkoutSession.PaymentIntent != nil {
			intentID = checkoutSession.PaymentIntent.ID
		} else if checkoutSession.Subscription != nil {
			intentID = checkoutSession.Subscription.ID
		} else {
			return errors.New("unrecognized response from Stripe")
		}

		productID, err := processOrder(intentID, refID, lineItems.LineItem())

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error w/ stripe order: %v\n", err)
			return err
		}

		// Send email (if applicable)
		email, err := db.GetUserEmailByPaymentID(refID)
		if err == nil && len(email) != 0 {
			err := mail.CreateOrderEmail(
				refID,
				stripeDescMap[productID],
				email,
			).Send()

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error sending order confirmation email")
			}
		}

		// Rotate user payment ID now that it's no longer needed
		err = db.RotateUserPaymentID(refID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rotating user payment ID: %v\n", err)
		}
	}
	return nil
}

// validateStripeEvent reads the request body and validates the contents of the
// request against the signature from the header. It returns the full Stripe
// event (if valid) and an error or nil.
func validateStripeEvent(payload []byte, sig string) (stripe.Event, error) {
	endpointSecret := os.Getenv("YEETFILE_STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, sig, endpointSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		return stripe.Event{}, err
	}

	return event, nil
}

// processOrder takes a Stripe payment intent ID, a customer reference ID, and
// an item purchased and updates the user's storage amount using the
// stripeProductAmounts mapping. Returns the product ID associated with the
// purchase and any error encountered.
func processOrder(
	intentID string,
	refID string,
	item *stripe.LineItem,
) (string, error) {
	productID := item.Price.Product.ID
	err := db.InsertNewOrder(intentID, refID, productID, int(item.Quantity))
	if err != nil {
		return "", err
	}

	// Check if this is a subscription vs a transfer upgrade
	if productID == stripeSubYearID || productID == stripeSubMonthID {
		var exp time.Time
		if productID == stripeSubYearID {
			// Add 1 year to exp
			exp = time.Now().AddDate(1, 0, 0)
		} else {
			// Add 1 month to exp
			exp = time.Now().AddDate(0, 1, 0)
		}

		err = db.SetUserMembershipExpiration(refID, exp)
		if err != nil {
			return "", err
		}

		return productID, nil
	}

	// Update user storage capacity
	amount, ok := stripeProductAmounts[productID]
	if ok {
		// TODO: Should storage be added regardless of db entry success?
		err = db.AddUserStorage(refID, amount*int(item.Quantity))
		if err != nil {
			return "", err
		}
	} else {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Unable to find product mapping for %s\n",
			item.Price.Product.ID)
	}

	return productID, nil
}

// init sets up the Stripe library with the developer's private key
func init() {
	stripe.Key = os.Getenv("YEETFILE_STRIPE_KEY")
	if len(stripe.Key) == 0 {
		stripeReady = false
	}

	for _, str := range stripeRequirements {
		if len(str) == 0 {
			stripeReady = false
			break
		}
	}

}
