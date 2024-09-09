package mail

import (
	"bytes"
	"text/template"
	"yeetfile/shared/endpoints"
)

type ChangeEmail struct {
	Domain   string
	Endpoint string
	ChangeID string
}

var changeEmailSubject = "YeetFile Email Change"
var changeEmailTemplate = template.Must(template.New("").Parse(
	"Hello,\n\nA request to change your YeetFile email was submitted for your " +
		"account.\n\nIf this was intentional, use the following link to " +
		"finish updating your email:\n\n" +
		"{{.Endpoint}}\n\n" +
		"If you are using the YeetFile command line app, enter the code " +
		"below into the prompt: {{.ChangeID}}."))

func SendEmailChangeNotification(email, changeID string) error {
	endpoint := endpoints.HTMLChangeEmail.Format(smtpConfig.CallbackDomain, changeID)
	change := ChangeEmail{
		Endpoint: endpoint,
		ChangeID: changeID,
	}

	var buf bytes.Buffer
	err := changeEmailTemplate.Execute(&buf, change)
	if err != nil {
		return err
	}

	body := buf.String()
	go sendEmail(email, changeEmailSubject, body)
	return nil
}
