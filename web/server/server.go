package server

import (
	"log"
	"net/http"
	"yeetfile/web/server/auth"
	"yeetfile/web/server/html"
	"yeetfile/web/server/misc"
	"yeetfile/web/server/payments"
	"yeetfile/web/server/session"
	"yeetfile/web/server/transfer"
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
		// Transfer (upload/download)
		{POST, "/u", AuthMiddleware(transfer.InitUploadHandler)},
		{POST, "/u/*/*", AuthMiddleware(transfer.UploadDataHandler)},
		{GET, "/d/*", transfer.DownloadHandler},
		{GET, "/d/*/*", transfer.DownloadChunkHandler},

		// Auth (signup, login/logout, account mgmt, etc)
		{GET, "/verify", auth.VerifyHandler},
		{GET, "/session", session.SessionHandler},
		{GET, "/logout", auth.LogoutHandler},
		{POST, "/login", auth.LoginHandler},
		{POST, "/signup", auth.SignupHandler},
		{GET | PUT, "/account", auth.AccountHandler},

		// Payments (Stripe, BTCPay)
		{POST, "/stripe", payments.StripeWebhook},
		{GET, "/checkout", payments.StripeCheckout},

		// HTML
		{GET, "/", html.HomePageHandler},
		{GET, "/upload", html.HomePageHandler},
		{GET, "/*", html.DownloadPageHandler},
		{GET, "/signup", html.SignupPageHandler},
		{GET, "/login", html.LoginPageHandler},
		{GET, "/faq", html.FAQPageHandler},

		// Misc
		{
			GET,
			"/static/*/*",
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

	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
