package middleware

import (
	"strings"

	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/captainthx/calorie/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(repo user.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "invalid credentials")
			c.Abort()
			return
		}
		u, err := repo.GetUserByToken(header[7:])
		if err != nil {
			response.Unauthorized(c, "invalid credentials")
			c.Abort()
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
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		u := userData.(*user.Users)
		if u.Role != user.Admin {
			response.Forbidden(c, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}
