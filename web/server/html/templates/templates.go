package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

type BaseTemplate struct {
	LoggedIn       bool
	Title          string
	SuccessMessage string
	ErrorMessage   string
	Javascript     []string
	CSS            []string
}

type Template struct {
	Base BaseTemplate
}

type LoginTemplate struct {
	Base  BaseTemplate
	Meter int
}

type AccountTemplate struct {
	Base          BaseTemplate
	Email         string
	Meter         int
	IsActive      bool
	PaymentID     string
	ExpString     string
	ReadableMeter string
}

type VerificationTemplate struct {
	Base  BaseTemplate
	Email string
}

const (
	UploadHTML       = "upload.html"
	DownloadHTML     = "download.html"
	VerificationHTML = "verify.html"
	SignupHTML       = "signup.html"
	LoginHTML        = "login.html"
	AccountHTML      = "account.html"
	FaqHTML          = "faq.html"
)

//go:embed *.html
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
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			templateList = append(templateList, path)
		}
		return err
	})

	templates = template.Must(template.ParseFS(HTML, templateList...))
	if err != nil {
		panic(err)
	}
}
