package btcpay

import (
	"strconv"
	"yeetfile/backend/db"
	"yeetfile/backend/server/subscriptions"
	"yeetfile/backend/utils"
	"yeetfile/shared/constants"
)

type Invoice struct {
	ManuallyMarked     bool   `json:"manuallyMarked"`
	OverPaid           bool   `json:"overPaid"`
	DeliveryID         string `json:"deliveryId"`
	WebhookID          string `json:"webhookId"`
	OriginalDeliveryID string `json:"originalDeliveryId"`
	IsRedelivery       bool   `json:"isRedelivery"`
	Type               string `json:"type"`
	Timestamp          int    `json:"timestamp"`
	StoreID            string `json:"storeId"`
	InvoiceID          string `json:"invoiceId"`
	Metadata           struct {
		Promo                                string `json:"promo"`
		OrderID                              string `json:"orderId"`
		ItemCode                             string `json:"itemCode"`
		ItemDesc                             string `json:"itemDesc"`
		BuyerEmail                           string `json:"buyerEmail"`
		Description                          string `json:"description"`
		ContactEmail                         string `json:"contact_email"`
		Quantity                             string `json:"quantity"`
		InvoiceAmount                        string `json:"invoice_amount"`
		InvoiceAmountMultiplyAdjustmentPromo string `json:"invoice_amount_multiply_adjustment_promo"`
	} `json:"metadata"`
}

// FinalizeInvoice finishes updating the user's account depending on what
// they purchased through BTCPay. Note that BTCPay order IDs are the same as
// user payment IDs.
func FinalizeInvoice(invoice Invoice) error {
	orderType := invoice.Metadata.ItemCode
	quantity, err := strconv.Atoi(invoice.Metadata.Quantity)
	if err != nil || quantity <= 0 {
		return nil
	}

	utils.LogStruct(invoice)

	product, err := subscriptions.GetProductByTag(orderType)
	if err != nil {
		return err
	}

	exp, err := subscriptions.GetSubscriptionExpiration(product.Duration, quantity)
	if err != nil {
		return err
	}

	err = db.SetUserSubscription(
		invoice.Metadata.OrderID,
		orderType,
		constants.SubMethodBTCPay,
		exp,
		product.StorageGBReal,
		product.SendGBReal)
	if err != nil {
		return err
	}

	return nil
}
