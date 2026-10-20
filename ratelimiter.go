package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	requestLimit = 5
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

func rateLimiter(next http.Handler) http.Handler {
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

		if client.requestsCount >= requestLimit {
			retryAfter := time.Until(client.windowEnd)
			mu.Unlock()

			w.Header().Set("Retry after",
				fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
			http.Error(w,
				"Rate limit exceeded.",
				http.StatusTooManyRequests)
			return
		}

		client.requestsCount++
		remaining := requestLimit - client.requestsCount
		mu.Unlock()

		w.Header().Set("RateLimit limit",
			fmt.Sprintf("%d", requestLimit))
		w.Header().Set("RateLimit remaining",
			fmt.Sprintf("%d", remaining))

		next.ServeHTTP(w, r)
	})
}
