package middleware_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/captainthx/calorie/backend/internal/middleware"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/captainthx/calorie/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type capturedRecord struct {
	level slog.Level
	msg   string
	attrs map[string]any
}

type captureHandler struct {
	records []capturedRecord
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.records = append(h.records, capturedRecord{
		level: r.Level,
		msg:   r.Message,
		attrs: attrs,
	})
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func newLoggedRouter(capture *captureHandler) *gin.Engine {
	router := gin.New()
	router.Use(middleware.RequestLogger(slog.New(capture)))
	return router
}

func TestRequestLoggerSkipsSuccessfulRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &captureHandler{}
	router := newLoggedRouter(capture)
	router.GET("/food-entries", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/food-entries", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if len(capture.records) != 0 {
		t.Fatalf("want no log records, got %d", len(capture.records))
	}
}

func TestRequestLoggerWarnsOnClientErrorsOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &captureHandler{}
	router := newLoggedRouter(capture)
	router.GET("/food-entries/:id", func(c *gin.Context) {
		u := &user.Users{}
		u.ID = 7
		c.Set("user", u)
		response.Forbidden(c, "forbidden")
	})

	req := httptest.NewRequest(http.MethodGet, "/food-entries/9", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if len(capture.records) != 1 {
		t.Fatalf("want 1 log record, got %d", len(capture.records))
	}
	record := capture.records[0]
	if record.level != slog.LevelWarn {
		t.Fatalf("want warn level, got %v", record.level)
	}
	if record.msg != "request rejected" {
		t.Fatalf("want message %q, got %q", "request rejected", record.msg)
	}
	if got := record.attrs["path"]; got != "/food-entries/:id" {
		t.Fatalf("want path %q, got %v", "/food-entries/:id", got)
	}
	if got := record.attrs["method"]; got != http.MethodGet {
		t.Fatalf("want method %q, got %v", http.MethodGet, got)
	}
	if got := record.attrs["request_id"]; got != "req-123" {
		t.Fatalf("want request_id %q, got %v", "req-123", got)
	}
	gotUserID, ok := record.attrs["user_id"].(uint64)
	if !ok {
		t.Fatalf("want uint64 user_id attr, got %T", record.attrs["user_id"])
	}
	if gotUserID != 7 {
		t.Fatalf("want user_id %d, got %d", 7, gotUserID)
	}
	if _, ok := record.attrs["error"]; ok {
		t.Fatal("did not expect error attr on 4xx log")
	}
}

func TestRequestLoggerLogsInternalErrorsOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &captureHandler{}
	router := newLoggedRouter(capture)
	router.GET("/food-entries", func(c *gin.Context) {
		response.InternalServerError(c, errors.New("load food entries: db error"))
	})

	req := httptest.NewRequest(http.MethodGet, "/food-entries", nil)
	req.Header.Set("X-Request-ID", "req-500")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if len(capture.records) != 1 {
		t.Fatalf("want 1 log record, got %d", len(capture.records))
	}
	record := capture.records[0]
	if record.level != slog.LevelError {
		t.Fatalf("want error level, got %v", record.level)
	}
	if record.msg != "request failed" {
		t.Fatalf("want message %q, got %q", "request failed", record.msg)
	}
	if got := record.attrs["error"]; got != "load food entries: db error" {
		t.Fatalf("want error attr %q, got %v", "load food entries: db error", got)
	}
	if _, ok := record.attrs["ip"]; !ok {
		t.Fatal("expected ip attr on 5xx log")
	}
}

func TestRequestLoggerSkipsOptionsAndDocs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	capture := &captureHandler{}
	router := newLoggedRouter(capture)
	router.OPTIONS("/food-entries", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/docs/*any", func(c *gin.Context) {
		response.NotFound(c, "not found")
	})

	optionsReq := httptest.NewRequest(http.MethodOptions, "/food-entries", nil)
	optionsW := httptest.NewRecorder()
	router.ServeHTTP(optionsW, optionsReq)

	docsReq := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
	docsW := httptest.NewRecorder()
	router.ServeHTTP(docsW, docsReq)

	if len(capture.records) != 0 {
		t.Fatalf("want no log records, got %d", len(capture.records))
	}
}
