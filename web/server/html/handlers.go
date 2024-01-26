package html

import (
	"fmt"
	"net/http"
	"time"
	"yeetfile/shared"
	"yeetfile/web/db"
	"yeetfile/web/server/html/templates"
	"yeetfile/web/server/session"
)

const ErrorHeader = "ErrorMsg"
const SuccessHeader = "SuccessMsg"

const OrderConfMsg = "Your order confirmation code " +
	"is \"%s\" -- if you don't have an email on file, please " +
	"write this down in case you need to contact YeetFile " +
	"about your order!"

// HomePageHandler returns the homepage html if not logged in, otherwise the
// upload page should be returned
func HomePageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.UploadHTML,
		templates.Template{Base: templates.BaseTemplate{
			LoggedIn:     session.IsValidSession(req),
			Title:        "Upload",
			ErrorMessage: w.Header().Get(ErrorHeader),
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
func DownloadPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.DownloadHTML,
		templates.Template{Base: templates.BaseTemplate{
			LoggedIn:     session.IsValidSession(req),
			Title:        "Download",
			ErrorMessage: w.Header().Get(ErrorHeader),
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
func SignupPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.SignupHTML,
		templates.Template{Base: templates.BaseTemplate{
			LoggedIn:     session.IsValidSession(req),
			Title:        "Create Account",
			ErrorMessage: w.Header().Get(ErrorHeader),
			Javascript:   []string{"auth.js", "signup.js"},
			CSS:          []string{"auth.css"},
		}},
	)

	handleError(w, err)
}

// LoginPageHandler returns the HTML page for logging in
func LoginPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.LoginHTML,
		templates.Template{Base: templates.BaseTemplate{
			LoggedIn:     session.IsValidSession(req),
			Title:        "Log In",
			ErrorMessage: w.Header().Get(ErrorHeader),
			Javascript:   []string{"auth.js"},
			CSS:          []string{"auth.css"},
		}},
	)

	handleError(w, err)
}

// AccountPageHandler returns the HTML page for a user managing their account
func AccountPageHandler(w http.ResponseWriter, req *http.Request, user db.User) {
	success := req.URL.Query().Get("success")
	conf := req.URL.Query().Get("confirmation")
	successMsg := ""
	errorMsg := ""
	if len(success) > 0 && success == "1" {
		successMsg = "Successfully updated account! "
		if len(conf) > 0 {
			paymentID, err := db.GetPaymentIDBySessionID(conf)
			if err == nil {
				successMsg += fmt.Sprintf(OrderConfMsg, paymentID)
			}
		}
	} else if len(success) > 0 && success == "0" {
		errorMsg = "Failed to update account!"
	}

	err := templates.ServeTemplate(
		w,
		templates.AccountHTML,
		templates.AccountTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:       session.IsValidSession(req),
				Title:          "My Account",
				ErrorMessage:   errorMsg,
				SuccessMessage: successMsg,
				Javascript:     nil,
				CSS:            []string{"account.css"},
			},
			Email:         user.Email,
			Meter:         user.Meter,
			PaymentID:     user.PaymentID,
			ExpString:     user.MemberExp.Format("2 Jan 2006"),
			IsActive:      time.Now().Before(user.MemberExp),
			ReadableMeter: shared.ReadableFileSize(user.Meter),
		},
	)

	handleError(w, err)
}

// VerifyPageHandler returns the HTML page for verifying the user's email
func VerifyPageHandler(w http.ResponseWriter, req *http.Request, email string) {
	err := templates.ServeTemplate(
		w,
		templates.VerificationHTML,
		templates.VerificationTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:     session.IsValidSession(req),
				Title:        "Verify",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript:   nil,
				CSS:          nil,
			},
			Email: email,
		},
	)

	handleError(w, err)
}

// FAQPageHandler returns the FAQ HTML page
func FAQPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.FaqHTML,
		templates.Template{
			Base: templates.BaseTemplate{
				LoggedIn:     session.IsValidSession(req),
				Title:        "FAQ",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript:   nil,
				CSS:          []string{"faq.css"},
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
