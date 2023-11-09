package server

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"yeetfile/web/server/auth"
	"yeetfile/web/server/html"
	"yeetfile/web/server/misc"
	"yeetfile/web/server/payments"
	"yeetfile/web/server/transfer"
)

const (
	POST   = http.MethodPost
	GET    = http.MethodGet
	PUT    = http.MethodPut
	DELETE = http.MethodDelete
)

// Run maps URL paths to handlers for the server and begins listening on the
// configured port.
func Run(port string, files embed.FS) {
	r := &router{
		routes: make(map[Route]http.HandlerFunc),
	}

	// Transfer (upload/download)
	r.AddRoute(POST, "/u", AuthMiddleware(transfer.InitUploadHandler))
	r.AddRoute(POST, "/u/*/*", AuthMiddleware(transfer.UploadDataHandler))
	r.AddRoute(GET, "/d/*", transfer.DownloadHandler)
	r.AddRoute(GET, "/d/*/*", transfer.DownloadChunkHandler)

	// Auth (signup, login/logout, account mgmt, etc)
	r.AddRoute(GET, "/verify", auth.VerifyHandler)
	r.AddRoute(GET, "/session", auth.SessionHandler)
	r.AddRoute(PUT, "/logout", auth.LogoutHandler)
	r.AddRoute(POST, "/login", auth.LoginHandler)
	r.AddRoute(POST, "/signup", LimiterMiddleware(auth.SignupHandler))

	// Payments (Stripe, BTCPay)
	r.AddRoute(POST, "/stripe", payments.StripeWebhook)

	// HTML
	r.AddRoute(GET, "/", html.HomePageHandler)
	r.AddRoute(GET, "/upload", html.HomePageHandler)
	r.AddRoute(GET, "/*", html.DownloadPageHandler)
	r.AddRoute(GET, "/signup", html.SignupPageHandler)
	r.AddRoute(GET, "/faq", html.FAQPageHandler)

	// Misc
	r.AddRoute(GET, "/static/*/*", misc.FileHandler(files))
	r.AddRoute(GET, "/wordlist", misc.WordlistHandler)
	r.AddRoute(GET, "/up", misc.UpHandler)

	addr := fmt.Sprintf("localhost:%s", port)
	log.Printf("Running on http://%s\n", addr)

	err := http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
