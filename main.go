package main

import (
	"fmt"
	"log"
	"net/http"
)

var requestCount int

func handler(w http.ResponseWriter, r *http.Request) {
	requestCount++

	fmt.Printf("Request %d", requestCount)
}

func main() {
	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
