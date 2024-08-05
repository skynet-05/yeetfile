package db

import (
	"errors"
)

func CreateNewStripeCustomer(customerID, paymentID string) error {
	s := `INSERT INTO stripe (customer_id, payment_id) VALUES ($1, $2)`
	_, err := db.Exec(s, customerID, paymentID)

	return err
}

func GetPaymentIDByStripeCustomerID(customerID string) (string, error) {
	rows, err := db.Query(`SELECT payment_id FROM stripe WHERE customer_id = $1`, customerID)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var paymentID string
		err = rows.Scan(&paymentID)
		if err != nil {
			return "", err
		}

		return paymentID, nil
	}

	return "", errors.New("unable to find payment ID")
}

func GetStripeCustomerIDByPaymentID(paymentID string) (string, error) {
	rows, err := db.Query(`SELECT customer_id FROM stripe WHERE payment_id = $1`, paymentID)
	if err != nil {
		return "", err
	}

	defer rows.Close()
	if rows.Next() {
		var customerID string
		err = rows.Scan(&customerID)
		if err != nil {
			return "", err
		}

		return customerID, nil
	}

	return "", errors.New("unable to find customer ID")
}
