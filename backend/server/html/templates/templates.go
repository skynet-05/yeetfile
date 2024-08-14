package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"yeetfile/backend/config"
	"yeetfile/backend/server/subscriptions"
)

type BaseTemplate struct {
	LoggedIn       bool
	Title          string
	Page           string
	SuccessMessage string
	ErrorMessage   string
	Javascript     []string
	CSS            []string
	Version        string
	Config         config.ServerConfig
}

type Template struct {
	Base BaseTemplate
}

type SignupTemplate struct {
	Base                   BaseTemplate
	ServerPasswordRequired bool
}

type LoginTemplate struct {
	Base  BaseTemplate
	Meter int
}

type VaultTemplate struct {
	Base             BaseTemplate
	FolderName       string
	StorageAvailable int
	StorageUsed      int
}

type AccountTemplate struct {
	Base                 BaseTemplate
	Email                string
	Meter                int
	IsActive             bool
	PaymentID            string
	ExpString            string
	StorageAvailable     string
	StorageUsed          string
	SendAvailable        string
	SendUsed             string
	IsYearly             bool
	IsStripeUser         bool
	StripeConfigured     bool
	BTCPayConfigured     bool
	BillingConfigured    bool
	SubscriptionTemplate subscriptions.SubscriptionTemplateValues
}

type VerificationTemplate struct {
	Base  BaseTemplate
	Email string
}

type ForgotPasswordTemplate struct {
	Base  BaseTemplate
	Email string
	Code  string
}

const (
	SendHTML         = "send.html"
	VaultHTML        = "vault.html"
	DownloadHTML     = "download.html"
	VerificationHTML = "verify.html"
	SignupHTML       = "signup.html"
	LoginHTML        = "login.html"
	AccountHTML      = "account.html"
	ForgotHTML       = "forgot.html"
	FaqHTML          = "faq.html"
)

//go:embed *.html
//go:embed items/*.html
var HTML embed.FS

var templates *template.Template

// ServeTemplate uses the name of a template and a generic struct and attempts to
// generate the template content using the provided struct. If there's an issue
// using the provided struct to fill out the template, an error will be thrown
// back to the caller.
func ServeTemplate[T any](w http.ResponseWriter, name string, fields T) error {
	err := templates.ExecuteTemplate(w, name, fields)
	if err != nil {
		return err
	}

	return nil
}

// init loads all HTML template files in the HTML embedded collection of files
func init() {
	// Load templates
	var templateList []string
	err := fs.WalkDir(HTML, ".", func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".html") {
			templateList = append(templateList, path)
		}
		return err
	})

	templates = template.Must(template.ParseFS(HTML, templateList...))
	if err != nil {
		panic(err)
	}
}
