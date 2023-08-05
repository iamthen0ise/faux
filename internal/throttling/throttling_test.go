package throttling

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThrottlingMiddleware(t *testing.T) {
	handler := ThrottlingMiddleware(100, 200)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, world!"))
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	start := time.Now()
	http.Get(server.URL)

	// We expect the request to take at least 100ms due to the throttling middleware.
	assert.True(t, time.Since(start) >= 100*time.Millisecond)
}

func TestRateLimitMiddleware(t *testing.T) {
	handler := RateLimitMiddleware(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, world!"))
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	// First request should pass.
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Second request should fail due to rate limiting.
	resp, err = http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
