package html

import (
	"log"
	"net/http"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/db"
	"yeetfile/backend/server/html/templates"
	"yeetfile/backend/server/session"
	"yeetfile/backend/server/subscriptions"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

const ErrorHeader = "ErrorMsg"
const SuccessHeader = "SuccessMsg"

const OrderConfMsg = "Your order confirmation code " +
	"is \"%s\" -- if you don't have an email on file, please " +
	"write this down in case you need to contact YeetFile " +
	"about your order!"

// VaultPageHandler returns the html template used for interacting with files
// (uploading, renaming, downloading, deleting) in the user's vault
func VaultPageHandler(w http.ResponseWriter, req *http.Request, userID string) {
	userStorage, _, err := db.GetUserStorage(userID)
	if err != nil {
		handleError(w, err)
		return
	}

	err = templates.ServeTemplate(
		w,
		templates.VaultHTML,
		templates.VaultTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:     session.IsValidSession(req),
				Title:        "Vault",
				Page:         "vault",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript: []string{
					"vault.js",
					"ponyfill.min.js",
				},
				CSS:    []string{"vault.css"},
				Config: config.YeetFileConfig,
			},
			StorageUsed:      userStorage.StorageUsed,
			StorageAvailable: userStorage.StorageAvailable,
		},
	)

	handleError(w, err)
}

func SharedVaultPageHandler(w http.ResponseWriter, req *http.Request, userID string) {

}

// SendPageHandler returns the html template used for sending files
func SendPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.SendHTML,
		templates.LoginTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:     session.IsValidSession(req),
				Title:        "Send",
				Page:         "send",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript: []string{
					"jszip.min.js",
					"share.js",
				},
				CSS:    []string{"upload.css"},
				Config: config.YeetFileConfig,
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
				"download.js",
			},
			CSS:    []string{"download.css"},
			Config: config.YeetFileConfig,
		}},
	)

	handleError(w, err)
}

// SignupPageHandler returns the HTML page for signing up for an account
func SignupPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.SignupHTML,
		templates.SignupTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:     session.IsValidSession(req),
				Title:        "Create Account",
				ErrorMessage: w.Header().Get(ErrorHeader),
				Javascript:   []string{"signup.js"},
				CSS:          []string{"auth.css"},
				Config:       config.YeetFileConfig,
			},
			ServerPasswordRequired: config.YeetFileConfig.PasswordHash != nil,
		},
	)

	handleError(w, err)
}

// LoginPageHandler returns the HTML page for logging in
func LoginPageHandler(w http.ResponseWriter, req *http.Request) {
	err := templates.ServeTemplate(
		w,
		templates.LoginHTML,
		templates.Template{
			Base: templates.BaseTemplate{
				LoggedIn:       session.IsValidSession(req),
				Title:          "Log In",
				SuccessMessage: w.Header().Get(SuccessHeader),
				ErrorMessage:   w.Header().Get(ErrorHeader),
				Javascript:     []string{"login.js"},
				CSS:            []string{"auth.css"},
				Config:         config.YeetFileConfig,
			},
		},
	)

	handleError(w, err)
}

// AccountPageHandler returns the HTML page for a user managing their account
func AccountPageHandler(w http.ResponseWriter, req *http.Request, userID string) {
	user, err := db.GetUserByID(userID)
	if err != nil {
		handleError(w, err)
		return
	}

	successMsg, errorMsg := generateAccountMessages(req)
	isYearly := req.URL.Query().Has("yearly")

	err = templates.ServeTemplate(
		w,
		templates.AccountHTML,
		templates.AccountTemplate{
			Base: templates.BaseTemplate{
				LoggedIn:       session.IsValidSession(req),
				Title:          "My Account",
				Page:           "account",
				ErrorMessage:   errorMsg,
				SuccessMessage: successMsg,
				Javascript:     []string{"account.js"},
				CSS:            []string{"account.css"},
				Config:         config.YeetFileConfig,
			},
			Email:            user.Email,
			PaymentID:        user.PaymentID,
			ExpString:        user.MemberExp.Format("2 Jan 2006"),
			IsActive:         time.Now().Before(user.MemberExp),
			SendAvailable:    shared.ReadableFileSize(user.SendAvailable),
			SendUsed:         shared.ReadableFileSize(user.SendUsed),
			StorageAvailable: shared.ReadableFileSize(user.StorageAvailable),
			StorageUsed:      shared.ReadableFileSize(user.StorageUsed),
			IsYearly:         isYearly,
			IsStripeUser:     user.SubscriptionMethod == constants.SubMethodStripe,
			StripeConfigured: config.YeetFileConfig.StripeBilling.Configured,
			BTCPayConfigured: config.YeetFileConfig.BTCPayBilling.Configured,
			BillingConfigured: config.YeetFileConfig.StripeBilling.Configured ||
				config.YeetFileConfig.BTCPayBilling.Configured,
			SubscriptionTemplate: subscriptions.TemplateValues,
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
				Config:       config.YeetFileConfig,
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
				Config:       config.YeetFileConfig,
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
				Config:       config.YeetFileConfig,
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
	} else if len(success) > 0 && success == "0" {
		errorMsg = "Failed to update account!"
	}

	return successMsg, errorMsg
}

func handleError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("template error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
