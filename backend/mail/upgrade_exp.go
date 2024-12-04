package mail

import (
	"bytes"
	"text/template"
)

type UpgradeExpirationEmail struct {
	Domain string
}

var upgradeExpirationSubject = "YeetFile upgrade expiration"
var upgradeExpirationBodyTemplate = template.Must(template.New("").Parse(
	"Your YeetFile upgrade is expiring in 1 week. To continue using the " +
		"features of your purchased upgrade, you will need to login to " +
		"{{.Domain}} to extend the length of your upgrade period.\n\n" +
		"If you don't want to renew your upgrade right away, you have " +
		"1 month after your upgrade expires to remove excess vault storage. " +
		"Passwords and existing YeetFile Send files will be " +
		"unaffected if your upgrade expires.\n\n- YeetFile Support"))

// SendUpgradeExpirationEmail notifies a group of users that their upgrade is
// expiring in a week from the current date.
func SendUpgradeExpirationEmail(to []string) error {
	var buf bytes.Buffer

	upgradeExpEmail := UpgradeExpirationEmail{
		Domain: smtpConfig.CallbackDomain,
	}

	_ = upgradeExpirationBodyTemplate.Execute(&buf, upgradeExpEmail)
	body := buf.String()

	// sendEmail can take a while to return, so we're calling it in the
	// background here.
	go sendBccEmail(upgradeExpirationSubject, body, to)
	return nil
}
