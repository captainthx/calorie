package middleware

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/captainthx/calorie/backend/internal/user"
	sentry "github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)

		c.Next()

		status := c.Writer.Status()
		rawPath := c.Request.URL.Path
		path := c.FullPath()
		if path == "" {
			path = rawPath
		}
		if shouldSkipRequestLog(c.Request.Method, rawPath) || status < http.StatusBadRequest {
			return
		}

		msg := "request rejected"
		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration", requestDuration(time.Since(start)),
			"request_id", requestID,
		}

		if userData, exists := c.Get("user"); exists {
			if u, ok := userData.(*user.Users); ok {
				attrs = append(attrs, "user_id", u.ID)
			}
		}

		if status >= http.StatusInternalServerError {
			msg = "request failed"
			if ip := c.ClientIP(); ip != "" {
				attrs = append(attrs, "ip", ip)
			}
			if lastErr := c.Errors.Last(); lastErr != nil {
				attrs = append(attrs, "error", lastErr.Error())
				sentry.CaptureException(lastErr.Unwrap())
			}
			logger.Error(msg, attrs...)
			return
		}

		logger.Warn(msg, attrs...)
	}
}

func shouldSkipRequestLog(method, path string) bool {
	if method == http.MethodOptions {
		return true
	}
	if path == "/ping" || path == "/health" {
		return true
	}
	return path == "/docs" || strings.HasPrefix(path, "/docs/")
}

func requestDuration(d time.Duration) string {
	if d >= time.Millisecond {
		return d.Truncate(time.Millisecond).String()
	}
	return d.Truncate(time.Microsecond).String()
}

func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
