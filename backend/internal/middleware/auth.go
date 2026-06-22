package middleware

import (
	"net/http"
	"strings"

	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(repo user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorize"})
			return
		}
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid Credentials"})
			return
		}
		u, err := repo.GetUserByToken(header[7:])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid Credentials"})
			return
		}
		c.Set("user", u)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userData, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorize"})
			return
		}
		u := userData.(*user.Users)
		if u.Role != user.Admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			return
		}
		c.Next()
	}
}
