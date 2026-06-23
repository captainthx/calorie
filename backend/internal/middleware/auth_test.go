package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/captainthx/calorie/backend/internal/middleware"
	"github.com/captainthx/calorie/backend/internal/user"
)

func TestAdminMiddlewareBlocksRegularUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", &user.Users{Role: user.User})

	middleware.AdminMiddleware()(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("want 403, got %d", w.Code)
	}
}

func TestAdminMiddlewareAllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user", &user.Users{Role: user.Admin})

	middleware.AdminMiddleware()(c)

	if w.Code == http.StatusForbidden {
		t.Error("admin should not get 403")
	}
}
