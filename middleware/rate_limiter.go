package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// IPRateLimiter holds the rate limiters for each IP address
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
}

// NewIPRateLimiter creates a new IPRateLimiter
func NewIPRateLimiter() *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
	}
}

// AddIP creates a new rate limiter for an IP address
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Limite de 100 requisições por dia
	limiter := golang.org/x/time/rate.New(rate.Every(24*time.Hour/100), 100)
	i.ips[ip] = limiter
	return limiter
}

// GetLimiter returns the rate limiter for an IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip)
	}

	return limiter
}

// RateLimitMiddleware applies rate limiting to a handler
func RateLimitMiddleware(next http.Handler) http.Handler {
	limiter := NewIPRateLimiter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if !ok {
			// Se não houver usuário, limita por IP
			ip := r.RemoteAddr
			if !limiter.GetLimiter(ip).Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
		} else {
			// Se houver usuário, limita por ID de usuário
			userID := string(user.ID)
			if !limiter.GetLimiter(userID).Allow() {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
