package html

import (
	"net/http"
	"yeetfile/web/server/html/templates"
)

// HomePageHandler returns the homepage html if not logged in, otherwise the
// upload page should be returned
func HomePageHandler(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.UploadHTML,
		templates.Template{LoggedIn: true},
	)
}

// DownloadPageHandler returns the HTML page for downloading a file
func DownloadPageHandler(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.DownloadHTML,
		templates.Template{LoggedIn: true},
	)
}

// SignupPageHandler returns the HTML page for signing up for an account
func SignupPageHandler(w http.ResponseWriter, req *http.Request) {
	// TODO: Signup html
}

// FAQPageHandler returns the FAQ HTML page
func FAQPageHandler(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.FaqHTML,
		templates.Template{LoggedIn: true},
	)
}
