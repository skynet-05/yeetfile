package mail

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
	"yeetfile/web/utils"
)

var config SMTPConfig

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
	if config == (SMTPConfig{}) {
		// SMTP hasn't been configured, ignore this request
		log.Printf("Attempted to send email, but SMTP hasn't been configured")
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(config.Host, config.Port, config.Address, config.Password)
	d.TLSConfig = &tls.Config{ServerName: config.Host}

	err := d.DialAndSend(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		log.Println("Failed to send email")
	}
}

func init() {
	portEnv := utils.GetEnvVar("YEETFILE_EMAIL_PORT", "")
	port, err := strconv.Atoi(portEnv)
	if err != nil {
		log.Printf("Unable to read YEETFILE_EMAIL_PORT "+
			"as int: \"%s\"", portEnv)
		log.Println("Skipping SMTP setup...")
		return
	}

	from := fmt.Sprintf("\"YeetFile\" <%s>", os.Getenv("YEETFILE_EMAIL_ADDR"))

	config = SMTPConfig{
		From:           from,
		Address:        os.Getenv("YEETFILE_EMAIL_ADDR"),
		Host:           os.Getenv("YEETFILE_EMAIL_HOST"),
		Port:           port,
		Password:       os.Getenv("YEETFILE_EMAIL_PW"),
		CallbackDomain: os.Getenv("YEETFILE_CALLBACK_DOMAIN"),
	}
}
