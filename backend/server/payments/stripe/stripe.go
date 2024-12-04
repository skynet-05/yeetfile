package stripe

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v78"
	billingsession "github.com/stripe/stripe-go/v78/billingportal/session"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/subscription"
	"github.com/stripe/stripe-go/v78/webhook"
	"log"
	"os"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
	"yeetfile/backend/mail"
	"yeetfile/backend/server/upgrades"
	"yeetfile/backend/utils"
	"yeetfile/shared/endpoints"
)

const (
	paymentIDKey  = "paymentID"
	productTagKey = "productTag"
)

// processInvoiceEvent receives an incoming Stripe "invoice.paid" event and
// updates the user's subscription expiration and monthly transfer limit
func processInvoiceEvent(event *stripe.EventData) error {
	log.Println("Incoming 'invoice.paid' event from Stripe")
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Raw, &invoice)
	if err != nil {
		log.Printf("Error unmarshalling invoice data: %v\n", err)
		return err
	}

	utils.LogStruct(invoice)

	stripeCustomer, err := db.GetStripeCustomerByCustomerID(invoice.Customer.ID)
	if err != nil {
		log.Println("Failed to get stripe customer by ID:", err)
		return err
	} else if len(stripeCustomer.PaymentID) == 0 || len(stripeCustomer.ProductTag) == 0 {
		// Initial checkout hasn't occurred yet
		log.Println("Payment ID is not populated for invoice")
		return nil
	}

	quantity, err := findInvoiceQuantity(invoice.Lines.Data)
	if err != nil {
		log.Println("Failed to find product ID from invoice:", err)
		return err
	}

	err = setUserSubscription(
		stripeCustomer.PaymentID,
		stripeCustomer.ProductTag,
		quantity)
	if err != nil {
		log.Println("Failed to set user subscription:", err)
		return err
	}

	return nil
}

// updateSubscriptionID receives an incoming Stripe subscription creation event
// and updates the sub_id value in the stripe table with the ID from the event.
// This allows querying the subscription later to determine if it's still active
// or if renewal is upcoming.
func updateSubscriptionID(event *stripe.EventData) error {
	var sub stripe.Subscription
	err := json.Unmarshal(event.Raw, &sub)
	if err != nil {
		log.Printf("Error parsing webhook JSON: %v\n", err)
		return err
	}

	err = db.SetSubscriptionID(sub.ID, sub.Customer.ID)
	return err
}

// processCheckoutEvent receives an incoming Stripe checkout event and converts the
// event into a subscription for the user
func processCheckoutEvent(event *stripe.EventData) error {
	log.Println("Incoming 'checkout.session.completed' event from Stripe")
	var checkoutSession stripe.CheckoutSession
	err := json.Unmarshal(event.Raw, &checkoutSession)
	if err != nil {
		log.Printf("Error parsing webhook JSON: %v\n", err)
		return err
	}

	params := &stripe.CheckoutSessionListLineItemsParams{
		Session: stripe.String(checkoutSession.ID),
	}

	productTag, ok := checkoutSession.Metadata[productTagKey]
	if !ok {
		log.Printf("Stripe checkout missing product tag!")
		return errors.New("missing product tag")
	}

	product, err := upgrades.GetUpgradeByTag(productTag)
	if err != nil {
		log.Printf("Error fetching product ID for stripe order: %v\n", err)
		return err
	}

	result := session.ListLineItems(params)
	for result.Next() {
		if result.LineItem().Price.UnitAmount < 0 {
			continue
		}

		userPaymentID := checkoutSession.ClientReferenceID

		if checkoutSession.Customer != nil {
			err = processNewSubscription(
				checkoutSession.ClientReferenceID,
				checkoutSession.Customer.ID)

			if err != nil {
				log.Printf("Error creating sub: %v\n", err)
				return err
			}

			err = db.SetProductTag(
				checkoutSession.ClientReferenceID,
				checkoutSession.Customer.ID,
				productTag)

			if err != nil {
				log.Printf("Error setting stripe product tag: %v\n", err)
				return err
			}

			customerParams := &stripe.CustomerParams{}
			customerParams.AddMetadata(
				"paymentID",
				checkoutSession.ClientReferenceID)
			_, err = customer.Update(
				checkoutSession.Customer.ID,
				customerParams)
			if err != nil {
				log.Printf("Failed to set payment ID for stripe customer")
			}
		}

		err = setUserSubscription(
			userPaymentID,
			product.Tag,
			int(result.LineItem().Quantity))
		if err != nil {
			return err
		}

		// Send email (if applicable)
		email, err := db.GetUserEmailByPaymentID(userPaymentID)
		if err == nil && len(email) != 0 {
			err = mail.CreateOrderEmail(
				product.Description,
				email,
			).Send()

			if err != nil {
				log.Println("Error sending confirmation email")
			}
		}
	}
	return nil
}

// IsActiveSubscription checks to see if the subscription matching the provided
// subID is active (has not been canceled and is not expired).
func IsActiveSubscription(subID string) (bool, error) {
	params := &stripe.SubscriptionParams{}
	result, err := subscription.Get(subID, params)
	if err != nil {
		return true, err
	}

	// Check for all conditions that qualify as having a canceled sub
	canceled := result.Status == stripe.SubscriptionStatusCanceled
	canceled = canceled || result.EndedAt > 0
	canceled = canceled || result.CanceledAt > 0 || result.CancelAt > 0

	return !canceled, nil
}

// DeleteCustomer deletes the specified customer from Stripe
func DeleteCustomer(customerID string) error {
	_, err := customer.Del(customerID, nil)
	return err
}

// GetCustomerPortalLink returns a customer portal for the user matching the
// specified payment ID.
func GetCustomerPortalLink(id string) (string, error) {
	customerID, err := db.GetStripeCustomerIDByPaymentID(id)
	if err != nil {
		return "", err
	}

	params := &stripe.BillingPortalSessionParams{
		Customer: stripe.String(customerID),
		ReturnURL: stripe.String(
			fmt.Sprintf(
				"%s/account",
				config.YeetFileConfig.Domain)),
	}

	result, err := billingsession.New(params)
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

// ProcessEvent receives an input stripe.Event and determines if/how a
// user's meter should be updated depending on the product they purchased.
func ProcessEvent(event stripe.Event) error {
	utils.LogStruct(event)

	// Currently only successful checkouts are handled by the webhook
	log.Println("Incoming ", event.Type)
	if event.Type == "invoice.paid" {
		return processInvoiceEvent(event.Data)
	} else if event.Type == "checkout.session.completed" {
		return processCheckoutEvent(event.Data)
	} else if event.Type == "customer.subscription.created" {
		return updateSubscriptionID(event.Data)
	}

	// Unsupported event, ignore...
	return nil
}

// findInvoiceQuantity takes a list of invoice line items and retrieves the
// quantity of the product ordered
func findInvoiceQuantity(items []*stripe.InvoiceLineItem) (int, error) {
	for _, item := range items {
		// Skip previous purchases that have been prorated
		if item.Amount < 0 {
			continue
		}

		return int(item.Quantity), nil
	}

	return 0, errors.New("no items in invoice")
}

// ValidateEvent reads the request body and validates the contents of the
// request against the signature from the header. It returns the full Stripe
// event (if valid) and an error or nil.
func ValidateEvent(payload []byte, sig string) (stripe.Event, error) {
	event, err := webhook.ConstructEventWithOptions(
		payload,
		sig,
		config.YeetFileConfig.StripeBilling.WebhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		return stripe.Event{}, err
	}

	return event, nil
}

// processNewSubscription takes a Stripe payment intent ID, a customer reference ID, and
// an item purchased and updates the user's storage amount using the
// stripeProductStorage mapping. Returns the product ID associated with the
// purchase and any error encountered.
func processNewSubscription(paymentID, customerID string) error {
	err := db.CreateNewStripeCustomer(customerID, paymentID, "")
	if err != nil {
		return err
	}

	return nil
}

// setUserSubscription retrieves values from storage/send/type maps and uses those
// to update the user's database entry
func setUserSubscription(paymentID, productID string, quantity int) error {
	product, err := upgrades.GetUpgradeByTag(productID)
	if err != nil {
		log.Printf("Error getting user subscription product '%s': %v\n", productID, err)
		return err
	}

	exp, err := upgrades.GetUpgradeExpiration(product.Duration, quantity)
	if err != nil {
		return err
	}

	err = db.SetUserUpgrade(
		paymentID,
		productID,
		exp,
		int64(product.StorageGB*1000*1000*1000),
		int64(product.SendGB*1000*1000*1000))
	if err != nil {
		log.Printf("Error setting user subscription: %v\n", err)
		return err
	}

	return nil
}

func generateProrationAmount(paymentID string) (int64, error) {
	subType, subExp, err := db.GetUserSubByPaymentID(paymentID)
	if err != nil {
		return 0, err
	}

	prevSubProration := float64(0)
	if len(subType) > 0 && subExp.After(time.Now()) {
		prevSub, err := upgrades.GetUpgradeByTag(subType)
		if err != nil {
			return 0, nil
		}

		prevSubProration = (float64(prevSub.Price) / float64(30)) *
			float64(utils.DayDiff(time.Now().UTC(), subExp))
	} else {
		return 0, nil
	}

	if prevSubProration <= 0 {
		return 0, nil
	}

	return int64(prevSubProration * 100), nil
}

func GenerateCheckoutLink(
	product upgrades.Upgrade,
	paymentID string,
	quantity int,
	baseURL string,
) (string, error) {
	successURL := endpoints.HTMLCheckoutComplete.Format(baseURL)
	cancelURL := endpoints.HTMLAccount.Format(baseURL)
	paymentMode := stripe.CheckoutSessionModePayment

	// Prorate previous subscriptions if applicable
	finalPrice := product.Price * 100 * int64(quantity)
	proratedAmount, err := generateProrationAmount(paymentID)
	if proratedAmount > 0 {
		proratedAmountPerItem := proratedAmount / int64(quantity)
		adjustedItemPrice := max((product.Price*100)-proratedAmountPerItem, 0)
		finalPrice = adjustedItemPrice
		product.Description += fmt.Sprintf(" [CREDITED $%.2f]", float64(proratedAmount/100))
	}

	priceData := &stripe.CheckoutSessionLineItemPriceDataParams{
		Currency: stripe.String("usd"),
		ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name:        stripe.String(product.Name),
			Description: stripe.String(product.Description),
		},
		UnitAmount: stripe.Int64(finalPrice),
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{{
		PriceData: priceData,
		Quantity:  stripe.Int64(int64(quantity)),
	}}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(paymentMode)),

		LineItems:  lineItems,
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			productTagKey: product.Tag,
			paymentIDKey:  paymentID,
		},
		ConsentCollection: &stripe.CheckoutSessionConsentCollectionParams{
			TermsOfService: stripe.String("required"),
		},
		ClientReferenceID:   stripe.String(paymentID),
		AllowPromotionCodes: stripe.Bool(true),
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}

func init() {
	stripe.Key = config.YeetFileConfig.StripeBilling.Key
}
