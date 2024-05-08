package mail

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
	"yeetfile/web/config"
)

var smtpConfig SMTPConfig

type SMTPConfig struct {
	From           string
	Address        string
	Host           string
	Port           int
	Password       string
	CallbackDomain string
}

// sendEmail sends an email to the address specified in the `to` arg, containing
// `subject` as the subject and `body` as the body.
func sendEmail(to string, subject string, body string) {
	if smtpConfig == (SMTPConfig{}) {
		// SMTP hasn't been configured, ignore this request
		log.Println("Attempted to send email, but SMTP hasn't been configured")
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(smtpConfig.Host, smtpConfig.Port, smtpConfig.Address, smtpConfig.Password)
	d.TLSConfig = &tls.Config{ServerName: smtpConfig.Host}

	err := d.DialAndSend(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		log.Println("Failed to send email")
	}
}

func init() {
	if !config.YeetFileConfig.Email.Configured {
		return
	}

	port, err := strconv.Atoi(config.YeetFileConfig.Email.Port)
	if err != nil {
		log.Printf("Unable to read email port as int: \"%s\"", config.YeetFileConfig.Email.Port)
		log.Println("Skipping SMTP setup...")
		return
	}

	from := fmt.Sprintf("\"YeetFile\" <%s>", config.YeetFileConfig.Email.Address)

	smtpConfig = SMTPConfig{
		From:           from,
		Address:        config.YeetFileConfig.Email.Address,
		Host:           config.YeetFileConfig.Email.Host,
		Port:           port,
		Password:       config.YeetFileConfig.Email.Password,
		CallbackDomain: config.YeetFileConfig.CallbackDomain,
	}
}
