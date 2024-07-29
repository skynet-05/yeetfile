package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"yeetfile/shared/endpoints"
	"yeetfile/web/server/auth"
	"yeetfile/web/server/html"
	"yeetfile/web/server/misc"
	"yeetfile/web/server/payments"
	"yeetfile/web/server/session"
	"yeetfile/web/server/transfer/send"
	"yeetfile/web/server/transfer/vault"
	"yeetfile/web/static"
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
		{GET, endpoints.DownloadVaultFileMetadata, AuthMiddleware(vault.DownloadHandler)},
		{GET, endpoints.DownloadVaultFileData, AuthMiddleware(vault.DownloadChunkHandler)},
		{ALL, endpoints.ShareFile, AuthMiddleware(vault.ShareHandler(false))},
		{ALL, endpoints.ShareFolder, AuthMiddleware(vault.ShareHandler(true))},
		//{GET, "/api/recycle-payment-id", AuthMiddleware(auth.RecyclePaymentIDHandler)},

		// Auth (signup, login/logout, account mgmt, etc)
		{GET, "/verify-email", auth.VerifyEmailHandler},
		{POST, endpoints.VerifyAccount, auth.VerifyAccountHandler},
		{GET, endpoints.Session, session.SessionHandler},
		{GET, endpoints.Logout, auth.LogoutHandler},
		{POST, endpoints.Login, auth.LoginHandler},
		{POST, endpoints.Signup, auth.SignupHandler},
		{GET | PUT, endpoints.Account, auth.AccountHandler},
		{GET | POST, endpoints.Forgot, auth.ForgotPasswordHandler},
		{POST, endpoints.Reset, auth.ResetPasswordHandler},
		{GET, endpoints.PubKey, LimiterMiddleware(AuthMiddleware(auth.PubKeyHandler))},

		// Payments (Stripe, BTCPay)
		{POST, "/webhook/stripe", payments.StripeWebhook},
		{GET, "/manage-sub", StripeMiddleware(AuthMiddleware(payments.StripeCustomerPortal))},
		{GET, "/checkout", StripeMiddleware(AuthMiddleware(payments.StripeCheckout))},
		{POST, "/webhook/btcpay", payments.BTCPayWebhook},
		{GET, "/checkout-btc", BTCPayMiddleware(AuthMiddleware(payments.BTCPayCheckout))},

		// HTML
		{GET, "/", html.SendPageHandler},
		{GET, "/account", AuthMiddleware(html.AccountPageHandler)},
		{GET, "/send", html.SendPageHandler},
		{GET, "/vault", AuthMiddleware(html.VaultPageHandler)},
		{GET, "/vault/*", AuthMiddleware(html.VaultPageHandler)},
		{GET, "/*", html.DownloadPageHandler},
		{GET, "/signup", html.SignupPageHandler},
		{GET, "/login", html.LoginPageHandler},
		{GET, "/faq", html.FAQPageHandler},

		// Misc
		{
			GET,
			"/static/*/?/*",
			misc.FileHandler("/static/", "", static.StaticFiles),
		},
		{GET, "/wordlist", misc.WordlistHandler},
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
