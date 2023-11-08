package mail

import (
	"crypto/tls"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
	"yeetfile/web/utils"
)

var config SMTPConfig

type SMTPConfig struct {
	From           string
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
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(config.Host, config.Port, config.From, config.Password)
	d.TLSConfig = &tls.Config{ServerName: config.Host}

	err := d.DialAndSend(m)
	if err != nil {
		log.Println("Failed to send verification email")
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

	config = SMTPConfig{
		From:           os.Getenv("YEETFILE_EMAIL_ADDR"),
		Host:           os.Getenv("YEETFILE_EMAIL_HOST"),
		Port:           port,
		Password:       os.Getenv("YEETFILE_EMAIL_PW"),
		CallbackDomain: os.Getenv("YEETFILE_EMAIL_CALLBACK_DOMAIN"),
	}
}
