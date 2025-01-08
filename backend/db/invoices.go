package db

import "time"

func AddInvoice(invoiceID, paymentID, source string) error {
	s := `INSERT INTO invoices
	      (invoice_id, payment_id, source, date)
	      VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(s, invoiceID, paymentID, source, time.Now().UTC())

	return err
}

func HasInvoice(invoiceID string) (bool, error) {
	var exists bool
	s := `SELECT EXISTS (SELECT 1 FROM invoices WHERE invoice_id = $1)`
	err := db.QueryRow(s, invoiceID).Scan(&exists)

	return exists, err
}
