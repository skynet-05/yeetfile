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
		// File Share
		{POST, endpoints.UploadSendFileMetadata, AuthMiddleware(send.UploadMetadataHandler)},
		{POST, endpoints.UploadSendFileData, AuthMiddleware(send.UploadDataHandler)},
		{POST, endpoints.UploadSendText, LimiterMiddleware(send.UploadPlaintextHandler)},
		{GET, endpoints.DownloadSendFileMetadata, send.DownloadHandler},
		{GET, endpoints.DownloadSendFileData, send.DownloadChunkHandler},

		// File Vault
		{ALL, endpoints.VaultFolder, AuthMiddleware(vault.FolderHandler)},
		{PUT | DELETE, endpoints.VaultFile, AuthMiddleware(vault.FileHandler)},
		{POST, endpoints.UploadVaultFileMetadata, AuthMiddleware(vault.UploadMetadataHandler)},
		{POST, endpoints.UploadVaultFileData, AuthMiddleware(vault.UploadDataHandler)},
		{GET, endpoints.DownloadVaultFileMetadata, AuthLimiterMiddleware(vault.DownloadHandler)},
		{GET, endpoints.DownloadVaultFileData, AuthMiddleware(vault.DownloadChunkHandler)},
		{ALL, endpoints.ShareFile, AuthMiddleware(vault.ShareHandler(false))},
		{ALL, endpoints.ShareFolder, AuthMiddleware(vault.ShareHandler(true))},
		//{GET, "/api/recycle-payment-id", AuthMiddleware(auth.RecyclePaymentIDHandler)},

		// Auth (signup, login/logout, account mgmt, etc)
		{POST, endpoints.VerifyEmail, auth.VerifyEmailHandler},
		{POST, endpoints.VerifyAccount, auth.VerifyAccountHandler},
		{GET, endpoints.Session, session.SessionHandler},
		{GET, endpoints.Logout, auth.LogoutHandler},
		{POST, endpoints.Login, auth.LoginHandler},
		{POST, endpoints.Signup, LimiterMiddleware(auth.SignupHandler)},
		{GET | PUT | DELETE, endpoints.Account, AuthMiddleware(auth.AccountHandler)},
		{POST, endpoints.Forgot, auth.ForgotPasswordHandler},
		{POST, endpoints.Reset, auth.ResetPasswordHandler},
		{GET, endpoints.PubKey, AuthLimiterMiddleware(auth.PubKeyHandler)},
		{GET, endpoints.ProtectedKey, AuthMiddleware(auth.ProtectedKeyHandler)},
		{PUT, endpoints.ChangePassword, AuthMiddleware(auth.ChangePasswordHandler)},

		// Payments (Stripe, BTCPay)
		{POST, endpoints.StripeWebhook, payments.StripeWebhook},
		{GET, endpoints.StripeManage, StripeMiddleware(AuthMiddleware(payments.StripeCustomerPortal))},
		{GET, endpoints.StripeCheckout, StripeMiddleware(AuthMiddleware(payments.StripeCheckout))},
		{POST, endpoints.BTCPayWebhook, payments.BTCPayWebhook},
		{GET, endpoints.BTCPayCheckout, BTCPayMiddleware(AuthMiddleware(payments.BTCPayCheckout))},

		// HTML
		{GET, endpoints.HTMLHome, html.SendPageHandler},
		{GET, endpoints.HTMLSend, html.SendPageHandler},
		{GET, endpoints.HTMLVault, AuthMiddleware(html.VaultPageHandler)},
		{GET, endpoints.HTMLVaultFolder, AuthMiddleware(html.VaultPageHandler)},
		{GET, endpoints.HTMLSendDownload, html.DownloadPageHandler},
		{GET, endpoints.HTMLSignup, html.SignupPageHandler},
		{GET, endpoints.HTMLLogin, html.LoginPageHandler},
		{GET, endpoints.HTMLForgot, html.ForgotPageHandler},
		{GET, endpoints.HTMLAccount, AuthMiddleware(html.AccountPageHandler)},
		{GET, endpoints.HTMLVerifyEmail, html.VerifyPageHandler},
		{GET, endpoints.HTMLChangePassword, AuthMiddleware(html.ChangePasswordPageHandler)},

		// Misc
		{
			GET,
			"/static/*/?/*",
			misc.FileHandler("/static/", "", static.StaticFiles),
		},
		{GET, "/up", misc.UpHandler},

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
