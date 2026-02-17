package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func FirebaseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := auth.FirebaseClient.VerifyIDToken(context.Background(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// ดึง uid
		c.Set("user_id", token.UID)

		// ดึง role จาก custom claim
		role, ok := token.Claims["role"].(string)
		if ok {
			c.Set("role", role)
		} else {
			c.Set("role", "user") // default
		}

		c.Next()
	}
}
