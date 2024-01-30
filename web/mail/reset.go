package mail

import (
	"bytes"
	"text/template"
)

type ResetEmail struct {
	Code   string
	Email  string
	Domain string
}

var resetSubject = "YeetFile Password Reset"
var resetBodyTemplate = template.Must(template.New("").Parse(
	"Your YeetFile password reset code is {{.Code}}.\n\n" +
		"Enter this code on the password reset page, or use the link " +
		"below to reset your password.\n\n" +
		"{{.Domain}}/forgot?email={{.Email}}&code={{.Code}}"))

// SendResetEmail formats and sends a password reset email to the user containing
// a verification code.
func SendResetEmail(code string, to string) error {
	var buf bytes.Buffer

	resetEmail := ResetEmail{
		Code:   code,
		Email:  to,
		Domain: config.CallbackDomain,
	}

	_ = resetBodyTemplate.Execute(&buf, resetEmail)
	body := buf.String()

	go sendEmail(to, resetSubject, body)
	return nil
}
