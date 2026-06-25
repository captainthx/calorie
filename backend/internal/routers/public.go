package routes

import (
	"net/http"

	"github.com/captainthx/calorie/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterPublicRoutes(rg gin.IRoutes, db *gorm.DB) {
	rg.GET("/ping", pingHandler)
	rg.GET("/health", healthHandler(db))
}

// Ping godoc
// @Summary Ping application
// @Description Return a simple pong response.
// @Tags Public
// @Produce json
// @Success 200 {object} response.SuccessBody{data=response.MessageData}
// @Router /ping [get]
func pingHandler(c *gin.Context) {
	response.Success(c, response.MessageData{Message: "pong"})
}

// Health godoc
// @Summary Health check
// @Description Return the API readiness based on database connectivity.
// @Tags Public
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /health [get]
func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}

		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
