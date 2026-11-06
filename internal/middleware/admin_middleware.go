package middleware

import (
	"test-golang-project/pkg/apperror"
	"test-golang-project/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware ensures that the authenticated user has is_admin = true
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("is_admin")
		if !exists {
			response.HandleError(c, apperror.Forbidden("Forbidden: you do not have permission to access this resource"))
			c.Abort()
			return
		}

		isAdmin, ok := val.(bool)
		if !ok || !isAdmin {
			response.HandleError(c, apperror.Forbidden("Forbidden: you do not have permission to access this resource"))
			c.Abort()
			return
		}

		c.Next()
	}
}
