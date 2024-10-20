package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"yeetfile/backend/server/auth"
	"yeetfile/backend/server/html"
	"yeetfile/backend/server/misc"
	"yeetfile/backend/server/payments"
	"yeetfile/backend/server/session"
	"yeetfile/backend/server/transfer/send"
	"yeetfile/backend/server/transfer/vault"
	"yeetfile/backend/static"
	"yeetfile/shared/endpoints"
)

type HttpMethod int

const (
	GET HttpMethod = 1 << iota
	PUT
	POST
	DELETE
	ALL = GET | PUT | POST | DELETE
)

var MethodMap = map[HttpMethod]string{
	GET:    http.MethodGet,
	PUT:    http.MethodPut,
	POST:   http.MethodPost,
	DELETE: http.MethodDelete,
}

// Run maps URL paths to handlers for the server and begins listening on the
// configured port.
func Run(addr string) {
	r := &router{
		routes: make(map[Route]http.HandlerFunc),
	}

	r.AddRoutes([]RouteDef{
		// YeetFile Send
		{POST, endpoints.UploadSendFileMetadata, AuthMiddleware(send.UploadMetadataHandler)},
		{POST, endpoints.UploadSendFileData, AuthMiddleware(send.UploadDataHandler)},
		{POST, endpoints.UploadSendText, LimiterMiddleware(send.UploadPlaintextHandler)},
		{GET, endpoints.DownloadSendFileMetadata, send.DownloadHandler},
		{GET, endpoints.DownloadSendFileData, send.DownloadChunkHandler},

		// YeetFile Vault
		{ALL, endpoints.VaultFolder, AuthMiddleware(vault.FolderHandler(vault.FileVault))},
		{GET | PUT | DELETE, endpoints.VaultFile, AuthMiddleware(vault.FileHandler)},
		{POST, endpoints.UploadVaultFileMetadata, AuthMiddleware(vault.UploadMetadataHandler)},
		{POST, endpoints.UploadVaultFileData, AuthMiddleware(vault.UploadDataHandler)},
		{GET, endpoints.DownloadVaultFileMetadata, AuthLimiterMiddleware(vault.DownloadHandler)},
		{GET, endpoints.DownloadVaultFileData, AuthMiddleware(vault.DownloadChunkHandler)},
		{ALL, endpoints.ShareFile, AuthMiddleware(vault.ShareHandler(false))},
		{ALL, endpoints.ShareFolder, AuthMiddleware(vault.ShareHandler(true))},

		// YeetFile Pass (YeetPass)
		{ALL, endpoints.PassFolder, AuthMiddleware(vault.FolderHandler(vault.PassVault))},
		{POST, endpoints.PassEntry, AuthMiddleware(vault.UploadMetadataHandler)},
		{DELETE, endpoints.PassEntry, AuthMiddleware(vault.FileHandler)},

		// Auth (signup, login/logout, account mgmt, etc)
		{POST, endpoints.VerifyEmail, auth.VerifyEmailHandler},
		{POST, endpoints.VerifyAccount, LimiterMiddleware(auth.VerifyAccountHandler)},
		{GET, endpoints.Session, session.SessionHandler},
		{GET, endpoints.Logout, auth.LogoutHandler},
		{GET | POST | DELETE, endpoints.TwoFactor, AuthMiddleware(auth.TwoFactorHandler)},
		{POST, endpoints.Login, LimiterMiddleware(auth.LoginHandler)},
		{POST, endpoints.Signup, LimiterMiddleware(auth.SignupHandler)},
		{GET | PUT | DELETE, endpoints.Account, AuthMiddleware(auth.AccountHandler)},
		{GET, endpoints.AccountUsage, AuthMiddleware(auth.AccountUsageHandler)},
		{POST, endpoints.Forgot, LimiterMiddleware(auth.ForgotPasswordHandler)},
		{GET, endpoints.PubKey, AuthLimiterMiddleware(auth.PubKeyHandler)},
		{GET, endpoints.ProtectedKey, AuthMiddleware(auth.ProtectedKeyHandler)},
		{POST | PUT, endpoints.ChangeEmail, AuthMiddleware(auth.ChangeEmailHandler)},
		{PUT, endpoints.ChangePassword, AuthMiddleware(auth.ChangePasswordHandler)},
		{POST, endpoints.ChangeHint, AuthMiddleware(auth.ChangeHintHandler)},
		{PUT, endpoints.RecyclePaymentID, AuthMiddleware(auth.RecyclePaymentIDHandler)},

		// Payments (Stripe, BTCPay)
		{POST, endpoints.StripeWebhook, payments.StripeWebhook},
		{GET, endpoints.StripeManage, StripeMiddleware(AuthMiddleware(payments.StripeCustomerPortal))},
		{GET, endpoints.StripeCheckout, StripeMiddleware(AuthMiddleware(payments.StripeCheckout))},
		{POST, endpoints.BTCPayWebhook, payments.BTCPayWebhook},
		{GET, endpoints.BTCPayCheckout, BTCPayMiddleware(AuthMiddleware(payments.BTCPayCheckout))},

		// HTML
		{GET, endpoints.HTMLHome, html.SendPageHandler},
		{GET, endpoints.HTMLSend, html.SendPageHandler},
		{GET, endpoints.HTMLPass, AuthMiddleware(html.PassVaultPageHandler)},
		{GET, endpoints.HTMLPassFolder, AuthMiddleware(html.PassVaultPageHandler)},
		{GET, endpoints.HTMLPassEntry, AuthMiddleware(html.PassVaultPageHandler)},
		{GET, endpoints.HTMLVault, AuthMiddleware(html.FileVaultPageHandler)},
		{GET, endpoints.HTMLVaultFolder, AuthMiddleware(html.FileVaultPageHandler)},
		{GET, endpoints.HTMLVaultFile, AuthMiddleware(html.FileVaultPageHandler)},
		{GET, endpoints.HTMLSendDownload, html.DownloadPageHandler},
		{GET, endpoints.HTMLSignup, NoAuthMiddleware(html.SignupPageHandler)},
		{GET, endpoints.HTMLLogin, NoAuthMiddleware(html.LoginPageHandler)},
		{GET, endpoints.HTMLForgot, NoAuthMiddleware(html.ForgotPageHandler)},
		{GET, endpoints.HTMLAccount, AuthMiddleware(html.AccountPageHandler)},
		{GET, endpoints.HTMLVerifyEmail, html.VerifyPageHandler},
		{GET, endpoints.HTMLChangeEmail, AuthMiddleware(html.ChangeEmailPageHandler)},
		{GET, endpoints.HTMLChangePassword, AuthMiddleware(html.ChangePasswordPageHandler)},
		{GET, endpoints.HTMLChangeHint, AuthMiddleware(html.ChangeHintPageHandler)},
		{GET, endpoints.HTMLTwoFactor, AuthMiddleware(html.TwoFactorPageHandler)},
		{GET, endpoints.HTMLServerInfo, html.ServerInfoPageHandler},
		{GET, endpoints.HTMLCheckoutComplete, html.CheckoutCompleteHandler},

		// Misc
		{ // Static folder files
			GET,
			"/static/*/?/*",
			misc.FileHandler("/static/", "", static.StaticFiles),
		},
		{ // Static subfolder files
			GET,
			"/static/*/?/*/*",
			misc.FileHandler("/static/", "", static.StaticFiles),
		},
		{GET, "/up", misc.UpHandler},
		{GET, endpoints.ServerInfo, misc.InfoHandler},

		// StreamSaver.js
		// These routes serve files directly from the stream_saver submodule
		{
			GET,
			"/mitm.html",
			misc.FileHandler("", "/stream_saver/", static.StreamSaverFiles),
		},
		{
			GET,
			"/StreamSaver.js",
			misc.FileHandler("", "/stream_saver/", static.StreamSaverFiles),
		},
		{
			GET,
			"/sw.js",
			misc.FileHandler("", "/stream_saver/", static.StreamSaverFiles),
		},
	})

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Running on http://%s\n", addr)
		if err := http.ListenAndServe(addr, r); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve returned err: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")
}
