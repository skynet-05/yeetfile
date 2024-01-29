package db

import (
	"errors"
	"log"
	"time"
)

func InsertNewBTCPayOrder(
	orderID string,
	invoiceID string,
	orderType string,
) error {
	s := `INSERT INTO btcpay (id, type, invoice_id, date)
	      VALUES ($1, $2, $3)`
	_, err := db.Exec(s, orderID, orderType, invoiceID, time.Now())

	return err
}

func GetBTCPayOrderTypeByID(orderID string) (string, error) {
	rows, err := db.Query(`
		SELECT type
		FROM btcpay
		WHERE id = $1`, orderID)
	if err != nil {
		log.Printf("Error querying for order type by ID: %v", err)
		return "", err
	}

	if rows.Next() {
		var orderType string
		err = rows.Scan(&orderType)
		if err != nil {
			return "", err
		}

		return orderType, nil
	}

	return "", errors.New("unable to find order type by id")
}
