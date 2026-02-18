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
			c.JSON(401, gin.H{"error": "Missing token"})
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

		emailClaims, ok := token.Claims["email"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email not verified"})
			c.Abort()
			return
		}

		email := emailClaims

		role := auth.GetRoleByEmail(email)

		// ดึง uid
		c.Set("user_id", token.UID)

		c.Set("email", email)
		c.Set("role", role)
		c.Set("user_id", token.UID)

		c.Next()
	}
}
