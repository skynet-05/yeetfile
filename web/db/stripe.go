package db

import (
	"errors"
	"log"
	"time"
)

func InsertNewStripeOrder(
	intentID string,
	paymentID string,
	productID string,
	sessionID string,
) error {
	s := `INSERT INTO stripe (intent_id, payment_id, product_id, session_id, date)
	      VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(s, intentID, paymentID, productID, sessionID, time.Now())

	return err
}

func GetStripePaymentIDBySessionID(sessionID string) (string, error) {
	rows, err := db.Query(`
		SELECT payment_id
		FROM stripe
		WHERE session_id = $1`, sessionID)
	if err != nil {
		log.Printf("Error querying for payment ID by session ID: %v", err)
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
