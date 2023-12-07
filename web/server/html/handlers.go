package html

import (
	"fmt"
	"net/http"
	"yeetfile/web/server/html/templates"
)

// HomePageHandler returns the homepage html if not logged in, otherwise the
// upload page should be returned
func HomePageHandler(w http.ResponseWriter, _ *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.UploadHTML,
		templates.Template{Base: templates.BaseTemplate{
			LoggedIn: false,
			Title:    "Upload",
			Javascript: []string{
				"jszip.min.js",
				"scrypt.min.js",
				"nacl-fast.min.js",
				"utils.js",
				"crypto.js",
				"upload.js",
			},
			CSS: []string{"upload.css"},
		}},
	)

	handleError(w, err)
}

// DownloadPageHandler returns the HTML page for downloading a file
func DownloadPageHandler(w http.ResponseWriter, _ *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.DownloadHTML,
		templates.Template{Base: templates.BaseTemplate{
			LoggedIn: false,
			Title:    "Download",
			Javascript: []string{
				"ponyfill.min.js",
				"scrypt.min.js",
				"nacl-fast.min.js",
				"crypto.js",
				"utils.js",
				"download.js",
			},
			CSS: []string{"download.css"},
		}},
	)

	handleError(w, err)
}

// SignupPageHandler returns the HTML page for signing up for an account
func SignupPageHandler(w http.ResponseWriter, _ *http.Request) {
	// TODO: Signup html
}

// VerifyPageHandler returns the HTML page for verifying the user's email
func VerifyPageHandler(w http.ResponseWriter, _ *http.Request, email string) {
	err := templates.ServeTemplate(
		w,
		templates.VerificationHTML,
		templates.VerifyTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:   false,
				Title:      "Verify",
				Javascript: nil,
				CSS:        nil,
			},
			Email: email,
		},
	)

	handleError(w, err)
}

// FAQPageHandler returns the FAQ HTML page
func FAQPageHandler(w http.ResponseWriter, _ *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.FaqHTML,
		templates.Template{
			Base: templates.BaseTemplate{
				LoggedIn:   false,
				Title:      "FAQ",
				Javascript: nil,
				CSS:        []string{"faq.css"},
			},
		},
	)

	handleError(w, err)
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		fmt.Printf("template error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
