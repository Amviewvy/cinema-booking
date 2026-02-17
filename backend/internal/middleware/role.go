package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleInterface, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no role"})
			c.Abort()
			return
		}

		role := roleInterface.(string)

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permission"})
			c.Abort()
			return
		}

		c.Next()
	}
}
