package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/config"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/util"
)

type contextKey string

const UserContextKey = contextKey("user")

// AuthMiddleware protects routes that require authentication
func AuthMiddleware(db *gorm.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		var user *models.User

		// Try to parse as JWT first
		claims := &util.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err == nil && token.Valid {
			// JWT is valid, find user by ID
			var u models.User
			if err := db.First(&u, claims.UserID).Error; err == nil {
				user = &u
			}
		} else {
			// If not a valid JWT, treat it as a FORGE_API_KEY
			var u models.User
			if err := db.First(&u, "forge_api_key = ?", tokenString).Error; err == nil {
				user = &u
			}
		}

		if user == nil {
			http.Error(w, "Invalid token or API key", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
