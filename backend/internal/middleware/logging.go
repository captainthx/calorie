package middleware

import (
	"log/slog"
	"time"

	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		if userData, exists := c.Get("user"); exists {
			if u, ok := userData.(*user.Users); ok {
				attrs = append(attrs, "user_id", u.ID, "user_role", u.Role)
			}
		}

		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			attrs = append(attrs, "errors", c.Errors.String())
			logger.Error("request completed", attrs...)
			return
		}

		logger.Info("request completed", attrs...)
	}
}
