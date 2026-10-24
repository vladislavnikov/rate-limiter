package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	rateLimiterPkg "rate-limiter"
)

func resetClients() {
	rateLimiterPkg.ResetClients()
}

func TestRateLimiterAllowsFiveRequestsAndRejectsSixth(t *testing.T) {
	resetClients()

	handler := rateLimiterPkg.RateLimiter(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d",
				i, rec.Code)
		}

		expectedRemaining := string(rune('5' - i))
		if got := rec.Header().Get("X-RateLimit-Remaining"); got != expectedRemaining {
			t.Errorf("request %d: expected remaining %s, got %s",
				i, expectedRemaining, got)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("request 6: expected 429, got %d", rec.Code)
	}
}

func TestDifferentClientsHaveSeparateLimits(t *testing.T) {
	resetClients()

	handler := rateLimiterPkg.RateLimiter(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	}))

	// First client uses all five requests.
	for i := 0; i < rateLimiterPkg.RequestLimit; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Second client should still be allowed.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.2:5678"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected second client to get 200, got %d",
			rec.Code)
	}
}
