package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

type BaseTemplate struct {
	LoggedIn   bool
	Title      string
	Javascript []string
	CSS        []string
}

type Template struct {
	Base BaseTemplate
}

type VerifyTemplate struct {
	Base  BaseTemplate
	Email string
}

const (
	UploadHTML       = "upload.html"
	DownloadHTML     = "download.html"
	VerificationHTML = "verify.html"
	FaqHTML          = "faq.html"
	FooterHTML       = "footer.html"
	HeaderHTML       = "header.html"
)

//go:embed *.html
var HTML embed.FS

var templates *template.Template

func ServeTemplate[T any](w http.ResponseWriter, name string, fields T) error {
	err := templates.ExecuteTemplate(w, name, fields)
	if err != nil {
		return err
	}

	return nil
}

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
