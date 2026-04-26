package middleware

import (
	"maps"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter enforces a per-IP token bucket limit across all HTTP requests.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rps      rate.Limit
	burst    int
	done     chan struct{}
}

// NewRateLimiter creates a RateLimiter that allows rps requests/sec with a burst of burst,
// and starts the background cleanup goroutine. Call Stop on shutdown.
func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rate.Limit(rps),
		burst:    burst,
		done:     make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// cleanup removes visitors not seen for more than 3 minutes, running every minute.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			maps.DeleteFunc(rl.visitors, func(_ string, v *visitor) bool {
				return time.Since(v.lastSeen) > 3*time.Minute
			})
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Stop signals the cleanup goroutine to exit.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// Middleware returns a handler wrapper that responds 429 when the per-IP bucket is empty.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if !rl.getLimiter(ip).Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"too many requests"}`)) //nolint:errcheck
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
