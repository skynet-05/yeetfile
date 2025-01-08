package stripe

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
	"yeetfile/backend/mail"
	"yeetfile/backend/server/upgrades"
	"yeetfile/backend/utils"
	"yeetfile/shared"
	"yeetfile/shared/endpoints"
)

const (
	paymentIDKey     = "paymentID"
	productTagKey    = "productTags"
	sendQuantityKey  = "sendQuantity"
	vaultQuantityKey = "vaultQuantity"
)

// processCheckoutEvent receives an incoming Stripe checkout event and converts the
// event into a subscription for the user
func processCheckoutEvent(event *stripe.EventData) error {
	log.Println("Incoming 'checkout.session.completed' event from Stripe")
	var (
		checkoutSession  stripe.CheckoutSession
		emailDescription string
	)

	err := json.Unmarshal(event.Raw, &checkoutSession)
	if err != nil {
		log.Printf("Error parsing webhook JSON: %v\n", err)
		return err
	}

	userPaymentID := checkoutSession.ClientReferenceID
	upgradeTags, ok := checkoutSession.Metadata[productTagKey]
	if !ok {
		log.Printf("Stripe checkout missing upgrade tag!")
		return errors.New("missing upgrade tag")
	}

	hasInvoice, err := db.HasInvoice(checkoutSession.ID)
	if err != nil || hasInvoice {
		log.Printf("Possible duplicate Stripe event (err: %v)\n", err)
		return err
	}

	splitTags := strings.Split(upgradeTags, ",")
	for _, upgradeTag := range splitTags {
		var upgrade shared.Upgrade
		upgrade, err = upgrades.GetUpgradeByTag(upgradeTag, upgrades.GetAllUpgrades())
		if err != nil {
			log.Printf("Error fetching upgrade ID for stripe order: %v\n", err)
			return err
		}

		var quantityStr string
		if upgrade.IsVaultUpgrade {
			quantityStr, ok = checkoutSession.Metadata[vaultQuantityKey]
		} else {
			quantityStr, ok = checkoutSession.Metadata[sendQuantityKey]
		}

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			quantity = 1
		}

		err = setUserSubscription(userPaymentID, upgrade.Tag, quantity)
		if err != nil {
			return err
		}

		if len(emailDescription) > 0 {
			emailDescription += "\n\n"
		}

		emailDescription += upgrade.Description
	}

	// Send email (if applicable)
	email, err := db.GetUserEmailByPaymentID(userPaymentID)
	if err == nil && len(email) != 0 {
		err = mail.CreateOrderEmail(emailDescription, email).Send()
		if err != nil {
			log.Println("Error sending confirmation email")
		}
	}

	err = db.AddInvoice(checkoutSession.ID, userPaymentID, "stripe")
	return err
}

// ProcessEvent receives an input stripe.Event and determines if/how a
// user's meter should be updated depending on the product they purchased.
func ProcessEvent(event stripe.Event) error {
	utils.LogStruct(event)

	// Currently only successful checkouts are handled by the webhook
	log.Println("Incoming ", event.Type)
	if event.Type == "checkout.session.completed" {
		return processCheckoutEvent(event.Data)
	}

	// Unsupported event, ignore...
	return nil
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

// setUserSubscription retrieves values from storage/send/type maps and uses those
// to update the user's database entry
func setUserSubscription(paymentID, productID string, quantity int) error {
	upgrade, err := upgrades.GetUpgradeByTag(productID, upgrades.GetAllUpgrades())
	if err != nil {
		log.Printf("Error getting user upgrade product '%s': %v\n", productID, err)
		return err
	}

	if upgrade.IsVaultUpgrade {
		var exp time.Time
		exp, err = upgrades.GetUpgradeExpiration(upgrade, quantity)
		if err != nil {
			return err
		}

		err = db.SetUserVaultUpgrade(
			paymentID,
			productID,
			exp,
			upgrade.Bytes)
	} else {
		err = db.SetUserSendUpgrade(
			paymentID,
			upgrade.Bytes*int64(quantity))
	}

	if err != nil {
		log.Printf("Error processing user upgrade: %v\n", err)
		return err
	}

	return nil
}

// generateProrationAmount uses the user's previous vault upgrade to modify a
// new upgrade's price based on unused months/years from their previous purchase.
func generateProrationAmount(paymentID string) (int64, error) {
	subType, subExp, err := db.GetUserSubByPaymentID(paymentID)
	if err != nil {
		return 0, err
	}

	prevSubProration := float64(0)
	if len(subType) > 0 && time.Now().Before(subExp) {
		prevSub, err := upgrades.GetUpgradeByTag(subType, upgrades.GetAllUpgrades())
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

// GenerateCheckoutLink generates a Stripe checkout link for the selected upgrade.
func GenerateCheckoutLink(
	upgrades []shared.Upgrade,
	paymentID string,
	baseURL string,
) (string, error) {
	var (
		tags          []string
		total         int64
		sendQuantity  int
		vaultQuantity int
		lineItems     []*stripe.CheckoutSessionLineItemParams
	)

	successURL := endpoints.HTMLCheckoutComplete.Format(baseURL)
	cancelURL := endpoints.HTMLAccount.Format(baseURL)
	paymentMode := stripe.CheckoutSessionModePayment

	for _, upgrade := range upgrades {
		finalPrice := upgrade.Price * 100 * int64(upgrade.Quantity)

		// Prorate previous vault upgrade if applicable
		if upgrade.IsVaultUpgrade {
			proratedAmount, err := generateProrationAmount(paymentID)
			if proratedAmount > 0 && err == nil {
				proratedAmountPerItem := proratedAmount / int64(upgrade.Quantity)
				adjustedItemPrice := max((upgrade.Price*100)-proratedAmountPerItem, 0)
				finalPrice = adjustedItemPrice
				upgrade.Description += fmt.Sprintf(" [CREDITED $%.2f]", float64(proratedAmount/100))
			}

			vaultQuantity = upgrade.Quantity
		} else {
			sendQuantity = upgrade.Quantity
		}

		priceData := &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String("usd"),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name:        stripe.String(upgrade.Name),
				Description: stripe.String(upgrade.Description),
			},
			UnitAmount: stripe.Int64(finalPrice / int64(upgrade.Quantity)),
		}

		lineItem := stripe.CheckoutSessionLineItemParams{
			PriceData: priceData,
			Quantity:  stripe.Int64(int64(upgrade.Quantity)),
			AdjustableQuantity: &stripe.CheckoutSessionLineItemAdjustableQuantityParams{
				Enabled: stripe.Bool(false),
			},
		}

		lineItems = append(lineItems, &lineItem)
		tags = append(tags, upgrade.Tag)
		total += finalPrice
	}

	// Calculate Stripe fees
	totalWithFees := calculateFees(total)
	fees := &stripe.CheckoutSessionLineItemParams{
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String("usd"),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name: stripe.String("Stripe Processing Fees"),
			},
			UnitAmount: stripe.Int64(totalWithFees),
		},
		Quantity: stripe.Int64(1),
	}

	lineItems = append(lineItems, fees)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(paymentMode)),

		LineItems:  lineItems,
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			productTagKey:    strings.Join(tags, ","),
			sendQuantityKey:  strconv.Itoa(sendQuantity),
			vaultQuantityKey: strconv.Itoa(vaultQuantity),
			paymentIDKey:     paymentID,
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

// calculateFees adjusts the total charge amount with Stripe's processing fees
// ($0.30 + 2.9% of the total)
func calculateFees(total int64) int64 {
	const (
		fixedFee                 int64 = 30 // $0.30
		percentageFeeNumerator   int64 = 29 // 2.9%
		percentageFeeDenominator int64 = 1000
	)

	denominator := percentageFeeDenominator - percentageFeeNumerator
	fees := (total*percentageFeeDenominator + fixedFee*percentageFeeDenominator) / denominator

	return fees - total
}

func init() {
	stripe.Key = config.YeetFileConfig.StripeBilling.Key
}
