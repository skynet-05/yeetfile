package mail

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
	"yeetfile/backend/config"
)

var smtpConfig SMTPConfig

type SMTPConfig struct {
	From           string
	User           string
	Host           string
	Port           int
	Password       string
	NoReply        string
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
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	send(m)
}

// sendBccEmail sends an email to the configured From address, but with the
// provided recipients included as Bcc recipients. This can be used to notify
// multiple users with the same message.
func sendBccEmail(subject, body string, recipients []string) {
	if smtpConfig == (SMTPConfig{}) {
		// SMTP hasn't been configured, ignore this request
		log.Println("Attempted to send email, but SMTP hasn't been configured")
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("To", smtpConfig.NoReply)
	m.SetHeader("Subject", subject)
	m.SetHeader("Bcc", recipients...)
	m.SetBody("text/plain", body)

	send(m)
}

func send(message *gomail.Message) {
	message.SetHeader("From", smtpConfig.From)

	d := gomail.NewDialer(
		smtpConfig.Host,
		smtpConfig.Port,
		smtpConfig.User,
		smtpConfig.Password)
	d.TLSConfig = &tls.Config{ServerName: smtpConfig.Host}

	err := d.DialAndSend(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		log.Println("Failed to send email")
	} else {
		log.Println("Email sent!")
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
		User:           config.YeetFileConfig.Email.User,
		Host:           config.YeetFileConfig.Email.Host,
		Port:           port,
		Password:       config.YeetFileConfig.Email.Password,
		NoReply:        config.YeetFileConfig.Email.NoReplyAddress,
		CallbackDomain: config.YeetFileConfig.Domain,
	}
}
