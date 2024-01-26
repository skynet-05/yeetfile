package db

import "time"

func InsertNewOrder(
	intentID string,
	paymentID string,
	productID string,
	quantity int,
) error {
	s := `INSERT INTO stripe (intent_id, payment_id, product_id, quantity, date)
	      VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(s, intentID, paymentID, productID, quantity, time.Now())

	return err
}
