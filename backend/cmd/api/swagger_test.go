package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func buildTestRouter(mode string) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()
	if mode != gin.ReleaseMode {
		swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
		r.GET("/docs/*any", func(c *gin.Context) {
			if c.Param("any") == "/" {
				c.Redirect(http.StatusTemporaryRedirect, "/docs/index.html")
				return
			}
			swaggerHandler(c)
		})
	}
	return r
}

func TestSwaggerGatedInReleaseMode(t *testing.T) {
	r := buildTestRouter(gin.ReleaseMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/docs/index.html", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("GET /docs/index.html in release mode: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSwaggerExposedInDebugMode(t *testing.T) {
	r := buildTestRouter(gin.DebugMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/docs/", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("GET /docs/ in debug mode: got %d, want %d", w.Code, http.StatusTemporaryRedirect)
	}
}
