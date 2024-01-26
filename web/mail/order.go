package mail

import (
	"bytes"
	"text/template"
)

type OrderEmail struct {
	RefID   string
	Product string
	Email   string
}

var orderSubject = "YeetFile Order Confirmation"
var orderBodyTemplate = template.Must(template.New("").Parse(
	"Thank you for using YeetFile! Your order summary is below.\n\n" +
		"- Product: {{.Product}}\n" +
		"- Order ID: {{.RefID}}\n\n" +
		"If you have any questions about your order, feel free to email " +
		"support@yeetfile.com or reply to this email."))

// CreateOrderEmail creates an OrderEmail struct for sending the order
// confirmation email
func CreateOrderEmail(refID string, desc string, email string) OrderEmail {
	return OrderEmail{
		RefID:   refID,
		Product: desc,
		Email:   email,
	}
}

// Send sends an order confirmation email to the user, containing the
// reference ID necessary for order inquiries.
func (o OrderEmail) Send() error {
	var buf bytes.Buffer

	_ = orderBodyTemplate.Execute(&buf, o)
	body := buf.String()

	// sendEmail can take a while to return, so we're calling it in the
	// background here.
	go sendEmail(o.Email, orderSubject, body)
	return nil
}
