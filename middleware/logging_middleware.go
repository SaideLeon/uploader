package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Pass to the next handler
		next.ServeHTTP(w, r)

		// Log request details
		logLine := ""
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if ok {
			logLine = "user_id=" + user.ID.String()
		} else {
			logLine = "user_id=anonymous"
		}

		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			logLine,
			time.Since(start),
		)
	})
}
