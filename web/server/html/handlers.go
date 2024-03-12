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
		templates.LoginTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:     session.IsValidSession(req),
				Title:        "Upload",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript: []string{
					"jszip.min.js",
					"utils.js",
					"crypto.js",
					"upload.js",
				},
				CSS: []string{"upload.css"},
			},
			Meter: 0,
		},
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
			LoggedIn:       session.IsValidSession(req),
			Title:          "Log In",
			SuccessMessage: w.Header().Get(SuccessHeader),
			ErrorMessage:   w.Header().Get(ErrorHeader),
			Javascript:     []string{"auth.js"},
			CSS:            []string{"auth.css"},
		}},
	)

	handleError(w, err)
}

// AccountPageHandler returns the HTML page for a user managing their account
func AccountPageHandler(w http.ResponseWriter, req *http.Request, user db.User) {
	successMsg, errorMsg := generateAccountMessages(req)

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
			Email:             user.Email,
			Meter:             user.Meter,
			PaymentID:         user.PaymentID,
			ExpString:         user.MemberExp.Format("2 Jan 2006"),
			IsActive:          time.Now().Before(user.MemberExp),
			ReadableMeter:     shared.ReadableFileSize(user.Meter),
			Membership3Months: shared.TypeSub3Months,
			Membership1Year:   shared.TypeSub1Year,
			Upgrade100GB:      shared.Type100GB,
			Upgrade500GB:      shared.Type500GB,
			Upgrade1TB:        shared.Type1TB,
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

// ForgotPageHandler returns the HTML page for resetting a user's password
func ForgotPageHandler(w http.ResponseWriter, req *http.Request, email string) {
	if len(email) == 0 {
		email = req.URL.Query().Get("email")
	}

	err := templates.ServeTemplate(
		w,
		templates.ForgotHTML,
		templates.ForgotPasswordTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:     false,
				Title:        "Forgot Password",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript:   nil,
				CSS:          nil,
			},
			Email: email,
			Code:  req.URL.Query().Get("code"),
		},
	)

	handleError(w, err)
}

// generateAccountMessages takes a request and generates success and error messages from
// the data contained in the request.
func generateAccountMessages(req *http.Request) (string, string) {
	success := req.URL.Query().Get("success")
	conf := req.URL.Query().Get("confirmation")
	fromBTC := req.URL.Query().Get("btcpay")

	successMsg := ""
	errorMsg := ""
	if len(success) > 0 && success == "1" {
		successMsg = "Successfully updated account! "

		if len(fromBTC) > 0 && fromBTC == "1" {
			successMsg += "BTCPay orders can take up to 5 minutes " +
				"to finalize. Your account will be updated once " +
				"your transaction has been validated. "
		}

		if len(conf) > 0 {
			paymentID, err := db.GetStripePaymentIDBySessionID(conf)
			if err == nil {
				successMsg += fmt.Sprintf(OrderConfMsg, paymentID)
			}
		}
	} else if len(success) > 0 && success == "0" {
		errorMsg = "Failed to update account!"
	}

	return successMsg, errorMsg
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		fmt.Printf("template error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
