package server

import (
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var visitors = make(map[string]*Visitor)
var mu sync.Mutex

func init() {
	go cleanup()
}

// getVisitor checks to see if an ip address is associated with a rate limiter,
// and returns it if so. If not, it creates a new entry in the visitors map
// associating the ip address with a new rate limiter.
func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	visitor, exists := visitors[ip]
	if !exists {
		limit := rate.Every(time.Second * 30)
		limiter := rate.NewLimiter(limit, 3)
		visitors[ip] = &Visitor{limiter, time.Now()}
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

		limiter := getVisitor(ip)
		if limiter.Allow() {
			next.ServeHTTP(w, req)
			return
		}

		w.WriteHeader(http.StatusTooManyRequests)
	}

	return handler
}

// AuthMiddleware enforces that a particular request has a valid session before
// handling. This is used primarily for the routes associated with uploading.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	handler := func(w http.ResponseWriter, req *http.Request) {
		// Skip auth if the app is in debug mode
		if os.Getenv("YEETFILE_DEBUG") == "1" {
			next.ServeHTTP(w, req)
			return
		}

		session, _ := GetSession(req)
		if ok, found := session.Values["auth"].(bool); !ok || !found {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// If the user is authenticated, call the next handler
		next.ServeHTTP(w, req)
	}

	return handler
}

// cleanup removes a ip->visitor pairing from the visitors map if they haven't
// repeated a request in over a minute.
func cleanup() {
	mu.Lock()
	for ip, v := range visitors {
		if time.Since(v.lastSeen) > time.Minute {
			delete(visitors, ip)
		}
	}
	mu.Unlock()

	time.Sleep(time.Second * 30)
	cleanup()
}
