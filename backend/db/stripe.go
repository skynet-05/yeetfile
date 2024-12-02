package db

import (
	"database/sql"
	"log"
	"time"
)

type StripeCustomerInfo struct {
	CustomerID string
	PaymentID  string
	SubID      string
	ProductTag string
	CreatedAt  time.Time
}

func CreateNewStripeCustomer(customerID, paymentID, subID string) error {
	existingCustomer, err := GetStripeCustomerByCustomerID(customerID)
	if err == sql.ErrNoRows {
		s := `INSERT INTO stripe
	              (customer_id, payment_id, sub_id, created_at)
	              VALUES ($1, $2, $3, $4)`
		_, err = db.Exec(s, customerID, paymentID, subID, time.Now().UTC())
	} else if paymentID != existingCustomer.PaymentID {
		if len(subID) > 0 {
			s := `UPDATE stripe SET payment_id=$1, sub_id=$2 WHERE customer_id=$3`
			_, err = db.Exec(s, paymentID, subID, customerID)
		} else {
			s := `UPDATE stripe SET payment_id=$1 WHERE customer_id=$2`
			_, err = db.Exec(s, paymentID, customerID)
		}
	}

	return err
}

func SetProductTag(paymentID, customerID, productTag string) error {
	s := `UPDATE stripe SET product_tag=$1 WHERE customer_id=$2 OR payment_id=$3`
	_, err := db.Exec(s, productTag, customerID, paymentID)
	if err != nil {
		return err
	}

	return nil
}

func SetSubscriptionID(subID, customerID string) error {
	s := `UPDATE stripe SET sub_id=$1 WHERE customer_id=$2`
	result, err := db.Exec(s, subID, customerID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == int64(0) {
		log.Println("Should create new customer with sub id", subID)
		return CreateNewStripeCustomer(customerID, "", subID)
	}

	return nil
}

func GetSubIDByPaymentID(paymentID string) (string, error) {
	var subID string
	s := `SELECT sub_id FROM stripe WHERE payment_id=$1`
	err := db.QueryRow(s, paymentID).Scan(&subID)

	return subID, err
}

// GetStripeCustomerByPaymentID returns the current Stripe customer's subscription
// ID, customer ID, and when their info was added to the database.
func GetStripeCustomerByPaymentID(paymentID string) (StripeCustomerInfo, error) {
	var (
		subID      string
		customerID string
		createdAt  time.Time
	)

	s := `SELECT sub_id, customer_id, created_at FROM stripe WHERE payment_id=$1`
	err := db.QueryRow(s, paymentID).Scan(&subID, &customerID, &createdAt)

	return StripeCustomerInfo{
		CustomerID: customerID,
		PaymentID:  paymentID,
		SubID:      subID,
		CreatedAt:  createdAt,
	}, err
}

func GetStripeCustomerByCustomerID(customerID string) (StripeCustomerInfo, error) {
	var (
		paymentID  string
		subID      string
		createdAt  time.Time
		productTag string
	)

	s := `SELECT payment_id, sub_id, created_at, product_tag FROM stripe WHERE customer_id = $1`
	err := db.QueryRow(s, customerID).Scan(&paymentID, &subID, &createdAt, &productTag)
	if err != nil {
		return StripeCustomerInfo{}, err
	}

	return StripeCustomerInfo{
		CustomerID: customerID,
		PaymentID:  paymentID,
		SubID:      subID,
		ProductTag: productTag,
		CreatedAt:  createdAt,
	}, nil
}

func GetStripeCustomerIDByPaymentID(paymentID string) (string, error) {
	var customerID string

	s := `SELECT customer_id FROM stripe WHERE payment_id=$1`
	err := db.QueryRow(s, paymentID).Scan(&customerID)
	if err != nil {
		return "", err
	}

	return customerID, nil
}
