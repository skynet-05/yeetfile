package mail

import (
	"bytes"
	"crypto/tls"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strconv"
	"text/template"
)

var config SMTPConfig

type SMTPConfig struct {
	From           string
	Host           string
	Port           int
	Password       string
	CallbackDomain string
}

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

// sendEmail sends an email to the address specified in the `to` arg, containing
// `subject` as the subject and `body` as the body.
func sendEmail(to string, subject string, body string) {
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

func init() {
	portEnv := os.Getenv("YEETFILE_EMAIL_PORT")
	port, err := strconv.Atoi(portEnv)
	if err != nil {
		log.Fatalf("Error reading YEETFILE_EMAIL_PORT "+
			"as int: \"%s\"", portEnv)
	}

	config = SMTPConfig{
		From:           os.Getenv("YEETFILE_EMAIL_ADDR"),
		Host:           os.Getenv("YEETFILE_EMAIL_HOST"),
		Port:           port,
		Password:       os.Getenv("YEETFILE_EMAIL_PW"),
		CallbackDomain: os.Getenv("YEETFILE_EMAIL_CALLBACK_DOMAIN"),
	}
}
