package stripe

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v78"
	billingsession "github.com/stripe/stripe-go/v78/billingportal/session"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
	"log"
	"os"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
	"yeetfile/backend/mail"
	"yeetfile/backend/server/subscriptions"
	"yeetfile/shared/constants"
)

var LinkMapping = map[string]string{
	subscriptions.MonthlyNovice: config.YeetFileConfig.StripeBilling.
		SubNoviceMonthlyLink,
	subscriptions.MonthlyRegular: config.YeetFileConfig.StripeBilling.
		SubRegularMonthlyLink,
	subscriptions.MonthlyAdvanced: config.YeetFileConfig.StripeBilling.
		SubAdvancedMonthlyLink,

	subscriptions.YearlyNovice: config.YeetFileConfig.StripeBilling.
		SubNoviceYearlyLink,
	subscriptions.YearlyRegular: config.YeetFileConfig.StripeBilling.
		SubRegularYearlyLink,
	subscriptions.YearlyAdvanced: config.YeetFileConfig.StripeBilling.
		SubAdvancedYearlyLink,
}

var stripeDescMap = map[string]string{
	config.YeetFileConfig.StripeBilling.SubNoviceMonthly: subscriptions.
		DescriptionMap[subscriptions.MonthlyNovice],
	config.YeetFileConfig.StripeBilling.SubRegularMonthly: subscriptions.
		DescriptionMap[subscriptions.MonthlyRegular],
	config.YeetFileConfig.StripeBilling.SubAdvancedMonthly: subscriptions.
		DescriptionMap[subscriptions.MonthlyAdvanced],

	config.YeetFileConfig.StripeBilling.SubNoviceYearly: subscriptions.
		DescriptionMap[subscriptions.YearlyNovice],
	config.YeetFileConfig.StripeBilling.SubRegularYearly: subscriptions.
		DescriptionMap[subscriptions.YearlyRegular],
	config.YeetFileConfig.StripeBilling.SubAdvancedYearly: subscriptions.
		DescriptionMap[subscriptions.YearlyAdvanced],
}

var stripeOrderTypeMap = map[string]string{
	config.YeetFileConfig.StripeBilling.SubNoviceMonthly: subscriptions.
		MonthlyNovice,
	config.YeetFileConfig.StripeBilling.SubRegularMonthly: subscriptions.
		MonthlyRegular,
	config.YeetFileConfig.StripeBilling.SubAdvancedMonthly: subscriptions.
		MonthlyAdvanced,

	config.YeetFileConfig.StripeBilling.SubNoviceYearly: subscriptions.
		YearlyNovice,
	config.YeetFileConfig.StripeBilling.SubRegularYearly: subscriptions.
		YearlyRegular,
	config.YeetFileConfig.StripeBilling.SubAdvancedYearly: subscriptions.
		YearlyAdvanced,
}

// stripeProductStorage maps product IDs to their respective amounts of storage
// that they grant a user
var stripeProductStorage = map[string]int{
	config.YeetFileConfig.StripeBilling.SubNoviceMonthly: subscriptions.
		StorageAmountMap[subscriptions.TypeNovice],
	config.YeetFileConfig.StripeBilling.SubNoviceYearly: subscriptions.
		StorageAmountMap[subscriptions.TypeNovice],

	config.YeetFileConfig.StripeBilling.SubRegularMonthly: subscriptions.
		StorageAmountMap[subscriptions.TypeRegular],
	config.YeetFileConfig.StripeBilling.SubRegularYearly: subscriptions.
		StorageAmountMap[subscriptions.TypeRegular],

	config.YeetFileConfig.StripeBilling.SubAdvancedMonthly: subscriptions.
		StorageAmountMap[subscriptions.TypeAdvanced],
	config.YeetFileConfig.StripeBilling.SubAdvancedYearly: subscriptions.
		StorageAmountMap[subscriptions.TypeAdvanced],
}

// stripeProductSend maps product IDs to their respective amounts of storage
// that they grant a user
var stripeProductSend = map[string]int{
	config.YeetFileConfig.StripeBilling.SubNoviceMonthly: subscriptions.
		SendAmountMap[subscriptions.TypeNovice],
	config.YeetFileConfig.StripeBilling.SubNoviceYearly: subscriptions.
		SendAmountMap[subscriptions.TypeNovice],

	config.YeetFileConfig.StripeBilling.SubRegularMonthly: subscriptions.
		SendAmountMap[subscriptions.TypeRegular],
	config.YeetFileConfig.StripeBilling.SubRegularYearly: subscriptions.
		SendAmountMap[subscriptions.TypeRegular],

	config.YeetFileConfig.StripeBilling.SubAdvancedMonthly: subscriptions.
		SendAmountMap[subscriptions.TypeAdvanced],
	config.YeetFileConfig.StripeBilling.SubAdvancedYearly: subscriptions.
		SendAmountMap[subscriptions.TypeAdvanced],
}

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

	paymentID, err := db.GetPaymentIDByStripeCustomerID(invoice.Customer.ID)
	if err != nil {
		return err
	} else if len(paymentID) == 0 {
		// Initial checkout hasn't occurred yet
		return nil
	}

	productID, err := FindInvoiceProductID(invoice.Lines.Data)
	if err != nil {
		return err
	}

	err = setUserSubscription(paymentID, productID)
	if err != nil {
		return err
	}

	return nil
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

	result := session.ListLineItems(params)
	for result.Next() {
		if result.LineItem().Price.UnitAmount < 0 {
			continue
		}

		userPaymentID := checkoutSession.ClientReferenceID
		productID, err := processNewSubscription(
			checkoutSession.ClientReferenceID,
			checkoutSession.Customer.ID,
			result.LineItem())

		if err != nil {
			log.Printf("Error w/ stripe order: %v\n", err)
			return err
		}

		// Send email (if applicable)
		email, err := db.GetUserEmailByPaymentID(userPaymentID)
		if err == nil && len(email) != 0 {
			err := mail.CreateOrderEmail(
				stripeDescMap[productID],
				email,
			).Send()

			if err != nil {
				log.Println("Error sending confirmation email")
			}
		}
	}
	return nil
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
				config.YeetFileConfig.CallbackDomain)),
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
	// Currently only successful checkouts are handled by the webhook
	if event.Type == "invoice.paid" {
		return processInvoiceEvent(event.Data)
	} else if event.Type == "checkout.session.completed" {
		return processCheckoutEvent(event.Data)
	}

	// Unsupported event, ignore...
	return nil
}

// FindInvoiceProductID takes a list of invoice line items and retrieves the
// ID of the product that the user has paid for, ignoring prorated prior items
// that can appear in the item list.
func FindInvoiceProductID(items []*stripe.InvoiceLineItem) (string, error) {
	for _, item := range items {
		// Skip previous purchases that have been prorated
		if item.Amount < 0 {
			continue
		}

		return item.Price.Product.ID, nil
	}

	return "", errors.New("no items in invoice")
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
func processNewSubscription(paymentID, customerID string, item *stripe.LineItem) (string, error) {
	productID := item.Price.Product.ID
	err := db.CreateNewStripeCustomer(customerID, paymentID)
	if err != nil {
		return "", err
	}

	err = setUserSubscription(paymentID, productID)
	if err != nil {
		return "", err
	}

	return productID, nil
}

// setUserSubscription retrieves values from storage/send/type maps and uses those
// to update the user's database entry
func setUserSubscription(paymentID, productID string) error {
	storage, storageOK := stripeProductStorage[productID]
	send, sendOK := stripeProductSend[productID]
	tag, tagOK := stripeOrderTypeMap[productID]
	if storageOK && sendOK && tagOK {
		exp, err := subscriptions.GetSubscriptionExpiration(tag)
		if err != nil {
			return err
		}

		err = db.SetUserSubscription(
			paymentID,
			tag,
			constants.SubMethodStripe,
			exp,
			storage,
			send)
		if err != nil {
			return err
		}
	} else {
		log.Printf("Error matching product %s storage/send/tag:\n"+
			"storage: %d (%v)\n"+
			"send: %d (%v)\n"+
			"tag: %s (%v)\n",
			productID,
			storage, storageOK,
			send, sendOK,
			tag, tagOK)
		return errors.New("missing required fields to update user subscription")
	}

	return nil
}

func init() {
	stripe.Key = config.YeetFileConfig.StripeBilling.Key
}
