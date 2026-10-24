package ratelimiter

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	RequestLimit = 5
	requestLimit = RequestLimit
	window       = time.Minute
)

type Client struct {
	requestsCount int
	windowEnd     time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client)
)

func ResetClients() {
	mu.Lock()
	defer mu.Unlock()
	clients = make(map[string]*Client)
}

func rateLimiter(next http.Handler) http.Handler {
	return RateLimiter(next)
}

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		mu.Lock()

		now := time.Now()
		client, exists := clients[ip]

		if !exists || now.After(client.windowEnd) {
			client = &Client{
				requestsCount: 0,
				windowEnd:     now.Add(window),
			}
			clients[ip] = client
		}

		if client.requestsCount >= RequestLimit {
			retryAfter := time.Until(client.windowEnd)
			mu.Unlock()

			w.Header().Set("Retry-After",
				fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
			http.Error(w,
				"Rate limit exceeded.",
				http.StatusTooManyRequests)
			return
		}

		client.requestsCount++
		remaining := RequestLimit - client.requestsCount
		mu.Unlock()

		w.Header().Set("X-RateLimit-Limit",
			fmt.Sprintf("%d", RequestLimit))
		w.Header().Set("X-RateLimit-Remaining",
			fmt.Sprintf("%d", remaining))

		next.ServeHTTP(w, r)
	})
}
