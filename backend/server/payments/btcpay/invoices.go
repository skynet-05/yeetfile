package btcpay

import (
	"log"
	"strconv"
	"time"
	"yeetfile/backend/db"
	"yeetfile/backend/server/upgrades"
	"yeetfile/backend/utils"
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

	hasInvoice, err := db.HasInvoice(invoice.InvoiceID)
	if err != nil || hasInvoice {
		log.Printf("Possible duplicate BTCPay invoice (err: %v)\n", err)
		return err
	}

	utils.LogStruct(invoice)

	upgrade, err := upgrades.GetUpgradeByTag(orderType, upgrades.GetAllUpgrades())
	if err != nil {
		return err
	}

	if upgrade.IsVaultUpgrade {
		var exp time.Time
		exp, err = upgrades.GetUpgradeExpiration(upgrade, quantity)
		if err != nil {
			return err
		}

		err = db.SetUserVaultUpgrade(
			invoice.Metadata.OrderID,
			orderType,
			exp,
			upgrade.Bytes)
	} else {
		err = db.SetUserSendUpgrade(
			invoice.Metadata.OrderID,
			upgrade.Bytes)
	}

	if err != nil {
		log.Println("Error processing BTCPay upgrade in database", err)
		return err
	}

	err = db.AddInvoice(invoice.InvoiceID, invoice.Metadata.OrderID, "btcpay")
	return err
}
