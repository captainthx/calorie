package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	registerFoodRoutes(rg, db)
}

func RegisterAdminRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	registerFoodAdminRoutes(rg, db)
}
