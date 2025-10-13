package middleware

import (
	"net/http"
	"vinyl-vault/internal/services"

	"github.com/gin-gonic/gin"
)

func RequireAdmin(userService *services.UserService) gin.HandlerFunc {

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		user, err := userService.GetUser(c.Request.Context(), userID.(uint64))
		if err != nil || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "super user access required"})
		}
		c.Next()
	}
}
