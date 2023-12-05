package templates

import (
	"embed"
	"html/template"
	"net/http"
)

type Template struct {
	LoggedIn bool
}

type VerifyTemplate struct {
	LoggedIn bool
	Email    string
}

var UploadHTML = "upload.html"
var DownloadHTML = "download.html"
var FaqHTML = "faq.html"
var VerifyHTML = "verify.html"
var FooterHTML = "footer.html"
var HeaderHTML = "header.html"

//go:embed *.html
var HTML embed.FS
var templates = template.Must(template.ParseFS(HTML,
	UploadHTML,
	DownloadHTML,
	FaqHTML,
	VerifyHTML,
	FooterHTML,
	HeaderHTML))

func ServeTemplate[T any](w http.ResponseWriter, name string, template T) {
	err := templates.ExecuteTemplate(w, name, template)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
