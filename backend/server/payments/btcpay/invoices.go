package btcpay

import (
	"errors"
	"log"
	"yeetfile/backend/db"
	"yeetfile/backend/server/subscriptions"
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
		InvoiceAmount                        string `json:"invoice_amount"`
		InvoiceAmountMultiplyAdjustmentPromo string `json:"invoice_amount_multiply_adjustment_promo"`
	} `json:"metadata"`
}

// FinalizeInvoice finishes updating the user's account depending on what
// they purchased through BTCPay
func FinalizeInvoice(invoice Invoice) error {
	// BTCPay order IDs are the same as user payment IDs
	orderType := invoice.Metadata.ItemCode
	storage, storageErr := subscriptions.GetSubscriptionStorage(orderType)
	send, sendErr := subscriptions.GetSubscriptionSend(orderType)
	if storageErr == nil && sendErr == nil {
		exp, err := subscriptions.GetSubscriptionExpiration(orderType)
		if err != nil {
			return err
		}

		err = db.SetUserSubscription(
			invoice.Metadata.OrderID,
			orderType,
			constants.SubMethodBTCPay,
			exp,
			storage,
			send)
		if err != nil {
			return err
		}
	} else {
		log.Printf("Error matching btcpay order %s storage/send:\n"+
			"storage: %d (%v)\n"+
			"send: %d (%v)\n"+
			orderType,
			storage, storageErr,
			send, sendErr)
		return errors.New("missing required fields to update user subscription")
	}

	return nil
}
