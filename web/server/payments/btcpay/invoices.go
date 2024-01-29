package btcpay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/mail"
)

var invoicesEndpoint = fmt.Sprintf("api/v1/stores/%s/invoices", storeID)

type NewInvoiceResponse struct {
	ID                                string   `json:"id"`
	StoreID                           string   `json:"storeId"`
	Amount                            string   `json:"amount"`
	CheckoutLink                      string   `json:"checkoutLink"`
	Status                            string   `json:"status"`
	AdditionalStatus                  string   `json:"additionalStatus"`
	MonitoringExpiration              int      `json:"monitoringExpiration"`
	ExpirationTime                    int      `json:"expirationTime"`
	CreatedTime                       int      `json:"createdTime"`
	AvailableStatusesForManualMarking []string `json:"availableStatusesForManualMarking"`
	Archived                          bool     `json:"archived"`
	Type                              string   `json:"type"`
	Currency                          string   `json:"currency"`
	Metadata                          struct {
	} `json:"metadata"`
	Checkout struct {
		SpeedPolicy           string   `json:"speedPolicy"`
		PaymentMethods        []string `json:"paymentMethods"`
		DefaultPaymentMethod  string   `json:"defaultPaymentMethod"`
		ExpirationMinutes     int      `json:"expirationMinutes"`
		MonitoringMinutes     int      `json:"monitoringMinutes"`
		PaymentTolerance      float64  `json:"paymentTolerance"`
		RedirectURL           any      `json:"redirectURL"`
		RedirectAutomatically bool     `json:"redirectAutomatically"`
		RequiresRefundEmail   any      `json:"requiresRefundEmail"`
		DefaultLanguage       any      `json:"defaultLanguage"`
		CheckoutType          any      `json:"checkoutType"`
		LazyPaymentMethods    any      `json:"lazyPaymentMethods"`
	} `json:"checkout"`
	Receipt struct {
		Enabled      any `json:"enabled"`
		ShowQR       any `json:"showQR"`
		ShowPayments any `json:"showPayments"`
	} `json:"receipt"`
}

type NewInvoiceRequest struct {
	Metadata struct {
		OrderID string `json:"orderId"`
	}
	Checkout struct {
		RedirectURL string `json:"redirectURL"`
	}
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type SettledPayment struct {
	AfterExpiration bool   `json:"afterExpiration"`
	PaymentMethod   string `json:"paymentMethod"`
	Payment         struct {
		ID           string `json:"id"`
		ReceivedDate int    `json:"receivedDate"`
		Value        string `json:"value"`
		Fee          string `json:"fee"`
		Status       string `json:"status"`
		Destination  string `json:"destination"`
	} `json:"payment"`
	DeliveryID         string `json:"deliveryId"`
	WebhookID          string `json:"webhookId"`
	OriginalDeliveryID string `json:"originalDeliveryId"`
	IsRedelivery       bool   `json:"isRedelivery"`
	Type               string `json:"type"`
	Timestamp          int    `json:"timestamp"`
	StoreID            string `json:"storeId"`
	InvoiceID          string `json:"invoiceId"`
	Metadata           struct {
		OrderID string `json:"orderId"`
	} `json:"metadata"`
}

type Invoice struct {
	ID           string
	CheckoutLink string
}

// GenerateBTCPayInvoice creates an invoice through BTCPay server and returns
// an Invoice struct containing the ID of the invoice and the checkout link
func GenerateBTCPayInvoice(paymentID string, price float32) (Invoice, error) {
	if !Ready {
		return Invoice{}, errors.New("BTCPay server not set up")
	}

	strPrice := fmt.Sprintf("%.2f", price)
	redirectURL := fmt.Sprintf(
		"%s/checkout?success=1&btcpay=1&confirmation={OrderId}",
		os.Getenv("YEETFILE_CALLBACK_DOMAIN"))

	newInvoice := NewInvoiceRequest{
		Metadata: struct {
			OrderID string `json:"orderId"`
		}(struct{ OrderID string }{OrderID: paymentID}),
		Checkout: struct {
			RedirectURL string `json:"redirectURL"`
		}(struct{ RedirectURL string }{RedirectURL: redirectURL}),
		Amount:   strPrice,
		Currency: "USD",
	}

	reqData, err := json.Marshal(newInvoice)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return Invoice{}, err
	}

	resp, err := sendRequest(http.MethodPost, invoicesEndpoint, reqData)
	decoder := json.NewDecoder(resp.Body)
	var invoice NewInvoiceResponse
	err = decoder.Decode(&invoice)
	if err != nil {
		return Invoice{}, err
	}

	return Invoice{
		ID:           invoice.ID,
		CheckoutLink: invoice.CheckoutLink,
	}, nil
}

// FinalizeInvoice finishes updating the user's account depending on what
// they purchased through BTCPay
func FinalizeInvoice(payment SettledPayment) error {
	// BTCPay order IDs are the same as user payment IDs
	orderID := payment.Metadata.OrderID
	orderType, err := db.GetBTCPayOrderTypeByID(orderID)
	if err != nil {
		return err
	}

	// Check if the order is for a membership
	if orderType == shared.TypeSub1Month || orderType == shared.TypeSub1Year {
		exp := shared.MembershipMap[orderType]()
		err = db.SetUserMembershipExpiration(orderID, exp)
		if err != nil {
			fmt.Printf("Error updating user expiration: %v\n", err)
			return err
		}
	} else {
		// Determine the amount of storage to be added to the user's account
		amount, ok := shared.UpgradeMap[orderType]
		if !ok {
			return errors.New("invalid order type")
		}

		err = db.AddUserStorage(orderID, amount)
		if err != nil {
			return err
		}
	}

	// Send an order confirmation email, if the user has an email registered
	email, _ := db.GetUserEmailByPaymentID(orderID)
	if len(email) > 0 {
		err := mail.CreateOrderEmail(orderID, shared.DescriptionMap[orderType], email).Send()
		if err != nil {
			return err
		}
	}
	_ = db.RotateUserPaymentID(orderID)

	return nil
}
