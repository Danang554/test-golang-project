package middleware

import (
	"fmt"
	"strings"
	"time"

	"test-golang-project/config"
	"test-golang-project/internal/repository"
	"test-golang-project/pkg/apperror"
	"test-golang-project/pkg/response"
	"test-golang-project/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config, revokedTokenRepo repository.RevokedTokenRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.HandleError(c, apperror.Unauthorized("Unauthorized. Please login first"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.HandleError(c, apperror.Unauthorized("Invalid authorization format. Format must be 'Bearer <token>'"))
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
			response.HandleError(c, apperror.Unauthorized("Invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.HandleError(c, apperror.Unauthorized("Invalid token claims"))
			c.Abort()
			return
		}

		sub, exists := claims["sub"]
		if !exists {
			response.HandleError(c, apperror.Unauthorized("Token does not contain user identification"))
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
			response.HandleError(c, apperror.Unauthorized("Invalid user identification format"))
			c.Abort()
			return
		}

		// Check if token has been revoked (logged out)
		if revokedTokenRepo != nil {
			tokenHash := utils.HashToken(tokenString)
			isRevoked, err := revokedTokenRepo.IsRevoked(c.Request.Context(), tokenHash)
			if err != nil {
				response.HandleError(c, apperror.Internal(err, "Failed to verify token revocation"))
				c.Abort()
				return
			}
			if isRevoked {
				response.HandleError(c, apperror.Unauthorized("Unauthorized. Please login first"))
				c.Abort()
				return
			}
		}

		// Extract token expiration and admin status
		var expTime time.Time
		if expVal, exists := claims["exp"]; exists {
			if expFloat, ok := expVal.(float64); ok {
				expTime = time.Unix(int64(expFloat), 0)
			}
		}

		var isAdmin bool
		if adminVal, exists := claims["is_admin"]; exists {
			if b, ok := adminVal.(bool); ok {
				isAdmin = b
			}
		}

		// Set context values for downstream handlers
		c.Set("user_id", userID)
		c.Set("is_admin", isAdmin)
		c.Set("token_string", tokenString)
		c.Set("token_exp", expTime)
		c.Next()
	}
}
