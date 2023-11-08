package mail

import (
	"bytes"
	"text/template"
)

type VerificationEmail struct {
	Code   string
	Email  string
	Domain string
}

var verificationSubject = "YeetFile Email Verification"
var verificationBodyTemplate = template.Must(template.New("").Parse(
	"Your YeetFile verification code is {{.Code}}.\n\n" +
		"Enter this code on the verification page, or use the link " +
		"below to finish verifying your account.\n\n" +
		"{{.Domain}}/verify?email={{.Email}}&code={{.Code}}"))

// SendVerificationEmail formats a standard verification email body using the
// code generated on signup and sends the email to the user.
func SendVerificationEmail(code string, to string) error {
	var buf bytes.Buffer

	verificationEmail := VerificationEmail{
		Code:   code,
		Email:  to,
		Domain: config.CallbackDomain,
	}

	_ = verificationBodyTemplate.Execute(&buf, verificationEmail)
	body := buf.String()

	// sendEmail can take a while to return, so we're calling it in the
	// background here.
	go sendEmail(to, verificationSubject, body)
	return nil
}
