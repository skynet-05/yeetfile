package templates

import (
	"embed"
	"html/template"
	"net/http"
)

type Template struct {
	LoggedIn bool
}

var UploadHTML = "upload.html"
var DownloadHTML = "download.html"
var FaqHTML = "faq.html"
var FooterHTML = "footer.html"

//go:embed *.html
var HTML embed.FS
var templates = template.Must(template.ParseFS(HTML,
	UploadHTML,
	DownloadHTML,
	FaqHTML,
	FooterHTML))

func ServeTemplate(w http.ResponseWriter, name string, template Template) {
	err := templates.ExecuteTemplate(w, name, template)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
