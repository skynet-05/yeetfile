package server

import (
	"golang.org/x/crypto/blake2b"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
	"yeetfile/backend/config"
	"yeetfile/backend/server/session"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var visitors = make(map[[32]byte]*Visitor)
var mu sync.Mutex

// getVisitor checks to see if an identifier (ip address or user id) is
// associated with a rate limiter, and returns it if so. If not, it creates a
// new entry in the visitors map associating the ip address with a new rate limiter.
func getVisitor(identifier string, path string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	idHash := blake2b.Sum256([]byte(identifier + path))
	visitor, exists := visitors[idHash]
	if !exists {
		limit := rate.Every(time.Second * constants.LimiterSeconds)
		limiter := rate.NewLimiter(limit, constants.LimiterAttempts)
		visitors[idHash] = &Visitor{limiter, time.Now()}
		return limiter
	}

	visitor.lastSeen = time.Now()
	return visitor.limiter
}

// LimiterMiddleware restricts requests to a particular route to prevent abuse
// of a handler function.
func LimiterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		ip := req.Header.Get("X-Forwarded-For")

		if len(ip) == 0 {
			fallbackIP, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ip = fallbackIP
		}

		limiter := getVisitor(ip, req.URL.Path)
		if limiter.Allow() {
			next.ServeHTTP(w, req)
			return
		}

		http.Error(
			w,
			"Too many requests from this IP address -- please wait and try again",
			http.StatusTooManyRequests)
	}

	return handler
}

// AuthMiddleware enforces that a particular request has a valid session before
// handling.
func AuthMiddleware(next session.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		// Skip auth if the app is in debug mode, otherwise validate session
		if session.IsValidSession(req) {
			// Call the next handler
			id, err := session.GetSessionAndUserID(req)
			if err != nil {
				return
			}

			next(w, req, id)
			return
		}

		http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
		return
	}

	return handler
}

// AuthLimiterMiddleware is like AuthMiddleware, but also restricts requests to
// the same constants.LimiterAttempts per constants.LimiterSeconds by session
// (unlike LimiterMiddleware which limits by IP address)
func AuthLimiterMiddleware(next session.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		// Skip auth if the app is in debug mode, otherwise validate session
		if session.IsValidSession(req) {
			// Call the next handler
			id, err := session.GetSessionAndUserID(req)
			if err != nil {
				return
			}

			limiter := getVisitor(id, req.URL.Path)
			if limiter.Allow() {
				next(w, req, id)
				return
			} else {
				http.Error(
					w,
					"Too many requests from this account -- please wait and try again",
					http.StatusTooManyRequests)
				return
			}
		}

		http.Redirect(w, req, string(endpoints.HTMLLogin), http.StatusTemporaryRedirect)
		return
	}

	return handler
}

// StripeMiddleware ensures that requests made to Stripe related endpoints are
// only processed if Stripe has been set up already.
func StripeMiddleware(next http.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		if !config.YeetFileConfig.StripeBilling.Configured {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		next(w, req)
	}

	return handler
}

// BTCPayMiddleware ensures that requests made to BTCPay related endpoints are
// only processed if BTCPay has been set up already.
func BTCPayMiddleware(next http.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		if !config.YeetFileConfig.BTCPayBilling.Configured {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		next(w, req)
	}

	return handler
}

// ManageLimiters removes an id->visitor pairing from the visitors map if they
// haven't repeated a limiter-enabled request in constants.LimiterSeconds.
func ManageLimiters() {
	mu.Lock()
	for ip, v := range visitors {
		if time.Since(v.lastSeen) > time.Minute {
			delete(visitors, ip)
		}
	}
	mu.Unlock()
}
