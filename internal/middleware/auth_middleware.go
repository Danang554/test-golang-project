package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"test-golang-project/config"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization format. Format must be 'Bearer <token>'", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token", nil)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "Invalid token claims", nil)
			c.Abort()
			return
		}

		sub, exists := claims["sub"]
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Token does not contain user identification", nil)
			c.Abort()
			return
		}

		// JSON numbers in MapClaims can decode as float64
		var userID uint
		switch v := sub.(type) {
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		case uint:
			userID = v
		default:
			response.Error(c, http.StatusUnauthorized, "Invalid user identification format", nil)
			c.Abort()
			return
		}

		// Set user ID into Gin context so controllers can access it
		c.Set("user_id", userID)
		c.Next()
	}
}
