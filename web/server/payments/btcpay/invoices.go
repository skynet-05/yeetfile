package btcpay

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// GenerateBTCPayInvoice creates an invoice through BTCPay server and returns
// the checkout link for that invoice.
func GenerateBTCPayInvoice(paymentID string, price float32) (string, error) {
	if !Ready {
		return "", errors.New("BTCPay server not set up")
	}

	strPrice := fmt.Sprintf("%f", price)
	orderID := fmt.Sprintf("btcpay_%s", paymentID)

	newInvoice := NewInvoiceRequest{
		Metadata: struct {
			OrderID string `json:"orderId"`
		}(struct{ OrderID string }{OrderID: orderID}),
		Amount:   strPrice,
		Currency: "USD",
	}

	reqData, err := json.Marshal(newInvoice)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", err
	}

	resp, err := sendRequest(http.MethodPost, invoicesEndpoint, reqData)
	decoder := json.NewDecoder(resp.Body)
	var invoice NewInvoiceResponse
	err = decoder.Decode(&invoice)
	if err != nil {
		return "", err
	}

	return invoice.CheckoutLink, nil
}
