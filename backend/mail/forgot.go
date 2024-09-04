package mail

import (
	"bytes"
	"strings"
	"text/template"
)

type ForgotPasswordEmail struct {
	Hint string
}

var forgotPasswordSubject = "YeetFile Password Hint"
var forgotPasswordTemplate = template.Must(template.New("").Parse(
	"Hello,\n\nYour YeetFile password hint was requested. If you did not " +
		"request this sent to you, please contact the YeetFile server " +
		"administrator.\n\n" +
		strings.Repeat("=", 80) + "\n\n" +
		"Your password hint is:\n" +
		"{{ .Hint }}"))

// SendPasswordHintEmail formats and sends a password hint email to the user.
func SendPasswordHintEmail(hint string, to string) error {
	var buf bytes.Buffer

	resetEmail := ForgotPasswordEmail{
		Hint: hint,
	}

	_ = forgotPasswordTemplate.Execute(&buf, resetEmail)
	body := buf.String()

	go sendEmail(to, forgotPasswordSubject, body)
	return nil
}
