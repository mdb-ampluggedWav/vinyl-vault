package middleware

import (
	"context"
	"net/http"
	"vinyl-vault/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {

	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func AdminRequired(userRepo services.UserRepository) gin.HandlerFunc {

	return func(c *gin.Context) {
		session := sessions.Default(c)

		isAdmin := session.Get("is_admin")
		if isAdmin == nil {

			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
				c.Abort()
				return
			}

			user, err := userRepo.FindByID(context.Background(), userID.(uint64))
			if err != nil || !user.IsAdmin {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				c.Abort()
				return
			}

			session.Set("is_admin", true)
			session.Save()
		}
		c.Next()
	}
}
