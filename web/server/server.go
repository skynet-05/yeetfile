package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
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
		{POST, "/send/u", AuthMiddleware(send.UploadMetadataHandler)},
		{POST, "/send/u/*/*", AuthMiddleware(send.UploadDataHandler)},
		{POST, "/send/plaintext", LimiterMiddleware(send.UploadPlaintextHandler)},
		{GET, "/send/d/*", send.DownloadHandler},
		{GET, "/send/d/*/*", send.DownloadChunkHandler},

		// File Vault
		{GET, "/api/vault", AuthMiddleware(vault.FolderViewHandler(true))},
		{GET, "/api/vault/*", AuthMiddleware(vault.FolderViewHandler(false))},
		{GET, "/api/shared", AuthMiddleware(vault.SharedFolderViewHandler(true))},
		{GET, "/api/shared/*", AuthMiddleware(vault.SharedFolderViewHandler(false))},
		{POST, "/api/vault/folder", AuthMiddleware(vault.NewFolderHandler)},
		{PUT | DELETE, "/api/vault/folder/*", AuthMiddleware(vault.ModifyFolderHandler)},
		{PUT | DELETE, "/api/vault/file/*", AuthMiddleware(vault.ModifyFileHandler)},
		{POST | DELETE, "/api/public/folder/*", AuthMiddleware(vault.PublicFolderHandler)},
		//{POST, "/api/public/file/*", AuthMiddleware(vault.PublicFileHandler)},
		{POST, "/api/vault/u", AuthMiddleware(vault.UploadMetadataHandler)},
		{POST, "/api/vault/u/*/*", AuthMiddleware(vault.UploadDataHandler)},
		{GET, "/api/vault/d/*", AuthMiddleware(vault.DownloadHandler)},
		{GET, "/api/vault/d/*/*", AuthMiddleware(vault.DownloadChunkHandler)},
		{GET | POST | PUT | DELETE, "/api/share/file/*", AuthMiddleware(vault.ShareHandler(false))},
		{GET | POST | PUT | DELETE, "/api/share/folder/*", AuthMiddleware(vault.ShareHandler(true))},
		//{GET, "/api/recycle-payment-id", AuthMiddleware(auth.RecyclePaymentIDHandler)},

		// Auth (signup, login/logout, account mgmt, etc)
		{GET, "/verify-email", auth.VerifyEmailHandler},
		{POST, "/verify-account", auth.VerifyAccountHandler},
		{GET, "/session", session.SessionHandler},
		{GET, "/logout", auth.LogoutHandler},
		{POST, "/login", auth.LoginHandler},
		{POST, "/signup", auth.SignupHandler},
		{GET | PUT, "/account", auth.AccountHandler},
		{GET | POST, "/forgot", auth.ForgotPasswordHandler},
		{POST, "/reset", auth.ResetPasswordHandler},
		{GET, "/pubkey", LimiterMiddleware(AuthMiddleware(auth.PubKeyHandler))},

		// Payments (Stripe, BTCPay)
		{POST, "/stripe", payments.StripeWebhook},
		{GET, "/checkout", AuthMiddleware(payments.StripeCheckout)},
		{POST, "/btcpay", payments.BTCPayWebhook},
		{GET, "/checkout-btc", AuthMiddleware(payments.BTCPayCheckout)},

		// HTML
		{GET, "/", html.SendPageHandler},
		{GET, "/send", html.SendPageHandler},
		{GET, "/vault", AuthMiddleware(html.VaultPageHandler)},
		{GET, "/vault/*", AuthMiddleware(html.VaultPageHandler)},
		{GET, "/shared", AuthMiddleware(html.SharedVaultPageHandler)},
		{GET, "/shared/*", AuthMiddleware(html.SharedVaultPageHandler)},
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
