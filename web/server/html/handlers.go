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
func SignupPageHandler(w http.ResponseWriter, _ *http.Request) {
	// TODO: Signup html
}

// VerifyPageHandler returns the HTML page for verifying the user's email
func VerifyPageHandler(w http.ResponseWriter, _ *http.Request, email string) {
	templates.ServeTemplate(
		w,
		templates.VerifyHTML,
		templates.VerifyTemplate{
			LoggedIn: true,
			Email:    email,
		},
	)
}

// FAQPageHandler returns the FAQ HTML page
func FAQPageHandler(w http.ResponseWriter, _ *http.Request) {
	templates.ServeTemplate(
		w,
		templates.FaqHTML,
		templates.Template{
			LoggedIn: true,
		},
	)
}
