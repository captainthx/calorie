package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/captainthx/calorie/backend/internal/config"
	"github.com/captainthx/calorie/backend/internal/food"
	"github.com/captainthx/calorie/backend/internal/middleware"
	routes "github.com/captainthx/calorie/backend/internal/routers"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/captainthx/calorie/backend/pkg/response"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "github.com/captainthx/calorie/backend/docs"
)

// @title Simple Calorie App API
// @version 1.0
// @description API for food entries, daily summaries, monthly price limits, and admin reports.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	logger := newLogger(os.Getenv("MODE"))
	if err := initTimezone(); err != nil {
		logger.Error("load timezone failed", "error", err)
		os.Exit(1)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}
	logger = newLogger(cfg.Mode)
	logger.Info("database connected")

	createUserRoleEnumQuery :=
		`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_enum') THEN
			CREATE TYPE user_role_enum AS ENUM ('ADMIN', 'USER');
		END IF;
	END $$`

	if err := cfg.Db.Exec(createUserRoleEnumQuery).Error; err != nil {
		logger.Error("create user role enum failed", "error", err)
		os.Exit(1)
	}

	cfg.Db.AutoMigrate(&user.Users{}, &food.FoodEntry{})

	// Seed users
	var count int64
	cfg.Db.Model(&user.Users{}).Count(&count)
	if count == 0 {
		cfg.Db.Create(&[]user.Users{
			{Name: "John", Role: user.User, Token: "user-token-123", DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000},
			{Name: "Jane", Role: user.User, Token: "user-token-456", DailyCalorieLimit: 2100, MonthlyPriceLimit: 1000},
			{Name: "Admin", Role: user.Admin, Token: "admin-token-789"},
		})
	}

	// Ensure default limits for users seeded before these columns existed (idempotent)
	cfg.Db.Model(&user.Users{}).
		Where("daily_calorie_limit = 0 AND role = ?", user.User).
		Updates(map[string]interface{}{"daily_calorie_limit": 2100, "monthly_price_limit": 1000})

	// Seed food entries
	var foodCount int64
	cfg.Db.Model(&food.FoodEntry{}).Count(&foodCount)
	if foodCount == 0 {
		var john, jane user.Users
		cfg.Db.Where("token = ?", "user-token-123").First(&john)
		cfg.Db.Where("token = ?", "user-token-456").First(&jane)
		seedFoodEntries(cfg.Db, john.ID, jane.ID)
	}

	gin.SetMode(cfg.Mode)
	if cfg.Mode == gin.DebugMode {
		gin.DebugPrintRouteFunc = func(string, string, string, int) {}
		gin.DebugPrintFunc = func(string, ...any) {}
	}
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Error("set trusted proxies failed", "error", err)
		os.Exit(1)
	}
	router.Use(gin.Recovery(), middleware.RequestLogger(logger))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     splitCSV(cfg.CORSAllowedOrigins),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/ping", func(c *gin.Context) {
		response.Success(c, response.MessageData{Message: "pong"})
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/docs/index.html")
	})
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	router.GET("/docs/*any", func(c *gin.Context) {
		if c.Param("any") == "/" {
			c.Redirect(http.StatusTemporaryRedirect, "/docs/index.html")
			return
		}
		swaggerHandler(c)
	})

	userRepo := user.NewUsersRepository(cfg.Db)

	api := router.Group("/api", middleware.AuthMiddleware(userRepo))
	admin := api.Group("/admin", middleware.AdminMiddleware())

	routes.RegisterRoutes(api, cfg.Db)
	routes.RegisterAdminRoutes(admin, cfg.Db)

	logger.Info("server starting", "port", cfg.Port, "mode", cfg.Mode)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func seedFoodEntries(db *gorm.DB, johnID, janeID uint) {
	entries := []food.FoodEntry{
		// John - yesterday 2200 cal (exceeds 2100 limit)
		{UserID: johnID, FoodName: "Breakfast", Calories: 700, Price: 50, EntryDate: ago(1, 8)},
		{UserID: johnID, FoodName: "Lunch", Calories: 800, Price: 75, EntryDate: ago(1, 12)},
		{UserID: johnID, FoodName: "Dinner", Calories: 700, Price: 80, EntryDate: ago(1, 18)},
		// John - today
		{UserID: johnID, FoodName: "Breakfast", Calories: 450, Price: 35, EntryDate: ago(0, 8)},
		{UserID: johnID, FoodName: "Lunch", Calories: 450, Price: 60, EntryDate: ago(0, 12)},
		// John - last 7 days
		{UserID: johnID, FoodName: "Rice & Curry", Calories: 600, Price: 45, EntryDate: ago(2, 12)},
		{UserID: johnID, FoodName: "Salad", Calories: 300, Price: 30, EntryDate: ago(3, 12)},
		{UserID: johnID, FoodName: "Pad Thai", Calories: 700, Price: 55, EntryDate: ago(4, 12)},
		{UserID: johnID, FoodName: "Tom Yum Soup", Calories: 400, Price: 65, EntryDate: ago(5, 12)},
		{UserID: johnID, FoodName: "Stir Fried Rice", Calories: 650, Price: 40, EntryDate: ago(6, 12)},
		// John - previous 7 days
		{UserID: johnID, FoodName: "Noodles", Calories: 700, Price: 50, EntryDate: ago(7, 12)},
		{UserID: johnID, FoodName: "Smoothie", Calories: 250, Price: 80, EntryDate: ago(8, 9)},
		{UserID: johnID, FoodName: "Sandwich", Calories: 500, Price: 45, EntryDate: ago(10, 12)},
		{UserID: johnID, FoodName: "Chicken Rice", Calories: 550, Price: 45, EntryDate: ago(12, 12)},
		{UserID: johnID, FoodName: "Pizza", Calories: 800, Price: 90, EntryDate: ago(13, 12)},
		// Jane - monthly price 1565 (exceeds 1000 limit)
		{UserID: janeID, FoodName: "Sushi", Calories: 600, Price: 280, EntryDate: ago(0, 13)},
		{UserID: janeID, FoodName: "Steak", Calories: 900, Price: 320, EntryDate: ago(1, 19)},
		{UserID: janeID, FoodName: "Lobster", Calories: 700, Price: 450, EntryDate: ago(2, 13)},
		{UserID: janeID, FoodName: "Salad", Calories: 200, Price: 35, EntryDate: ago(3, 12)},
		{UserID: janeID, FoodName: "Pasta", Calories: 500, Price: 75, EntryDate: ago(4, 12)},
		{UserID: janeID, FoodName: "Breakfast", Calories: 350, Price: 40, EntryDate: ago(5, 8)},
		{UserID: janeID, FoodName: "Thai food", Calories: 550, Price: 65, EntryDate: ago(6, 12)},
		{UserID: janeID, FoodName: "Dim sum", Calories: 600, Price: 120, EntryDate: ago(8, 12)},
		{UserID: janeID, FoodName: "Burger", Calories: 700, Price: 85, EntryDate: ago(11, 12)},
		{UserID: janeID, FoodName: "Ramen", Calories: 650, Price: 95, EntryDate: ago(12, 12)},
	}
	db.Create(&entries)
}

// ago returns start-of-day minus n days, at the given hour
func ago(n, hour int) time.Time {
	now := time.Now()
	d := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	return d.AddDate(0, 0, -n)
}

func initTimezone() error {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return fmt.Errorf("load timezone Asia/Bangkok: %w", err)
	}
	time.Local = ict
	return nil
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func newLogger(mode string) *slog.Logger {
	level := slog.LevelInfo
	if mode == gin.DebugMode || mode == "test" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	if mode == gin.ReleaseMode {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
