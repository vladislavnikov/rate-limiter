package ratelimiter

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(
		":8080",
		RateLimiter(mux),
	))
}
