package throttling

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/juju/ratelimit"
)

func ThrottlingMiddleware(low, high int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate a random delay between low and high.
			delay := low + rand.Intn(high-low+1)

			// Sleep for the delay duration.
			time.Sleep(time.Duration(delay) * time.Millisecond)

			// Call the next handler.
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitMiddleware(rps float32) func(http.Handler) http.Handler {
	// If rps is zero, just return the next handler without rate limiting
	if rps == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Else, implement rate limiting
	bucket := ratelimit.NewBucket(time.Minute/time.Duration(rps), int64(rps))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bucket.TakeAvailable(1) == 0 {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
