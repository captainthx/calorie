package routes

import (
	"github.com/captainthx/calorie/backend/internal/food"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerFoodRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	foodRepo := food.NewFoodRepository(db)
	userRepo := user.NewUsersRepository(db)
	foodSvc := food.NewFoodService(foodRepo, userRepo)
	foodHdl := food.NewHandler(foodSvc)

	rg.GET("/food-entries", foodHdl.List)
	rg.POST("/food-entries", foodHdl.Create)
	rg.PUT("/food-entries/:id", foodHdl.FullUpdate)
	rg.PATCH("/food-entries/:id", foodHdl.Update)
	rg.DELETE("/food-entries/:id", foodHdl.Delete)
	rg.GET("/daily-summary", foodHdl.DailySummary)
	rg.GET("/daily-summary-range", foodHdl.DailySummaryRange)
}

func registerFoodAdminRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	foodRepo := food.NewFoodRepository(db)
	userRepo := user.NewUsersRepository(db)
	foodSvc := food.NewFoodService(foodRepo, userRepo)
	foodAdmHdl := food.NewAdminHandler(foodSvc)

	rg.GET("/food-entries", foodAdmHdl.ListAll)
	rg.GET("/food-entries/:id", foodAdmHdl.GetByID)
	rg.POST("/food-entries", foodAdmHdl.Create)
	rg.PUT("/food-entries/:id", foodAdmHdl.FullUpdate)
	rg.PATCH("/food-entries/:id", foodAdmHdl.Update)
	rg.DELETE("/food-entries/:id", foodAdmHdl.Delete)
	rg.GET("/reports", foodAdmHdl.Report)
}
