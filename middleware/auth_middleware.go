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

		var tokenString string
		var user *models.User

		// Check for Bearer token format
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			// Try to parse as JWT
			claims := &util.Claims{}
			token, jwtErr := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.AppConfig.JWTSecret), nil
			})

			if jwtErr == nil && token.Valid {
				// JWT is valid, find user by ID
				var u models.User
				if err := db.First(&u, claims.UserID).Error; err == nil {
					user = &u
				}
			} else {
				// If JWT is invalid, try to use the tokenString as API key
				var u models.User
				if err := db.First(&u, "forge_api_key = ?", tokenString).Error; err == nil {
					user = &u
				}
			}
		} else {
			// If not "Bearer", assume it might be a raw API key
			tokenString = authHeader
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
