package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/captainthx/calorie/backend/internal/config"
	"github.com/captainthx/calorie/backend/internal/migrations"
	"github.com/captainthx/calorie/backend/internal/middleware"
	routes "github.com/captainthx/calorie/backend/internal/routers"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-gormigrate/gormigrate/v2"
	sentry "github.com/getsentry/sentry-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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
	initTimezone()
	initSentry()
	defer sentry.Flush(2 * time.Second)

	cfg, err := config.LoadConfig()
	logger := newLogger(os.Getenv("GIN_MODE"))
	if err != nil {
		logger.Error("load config failed", "error", err)
		os.Exit(1)
	}
	logger = newLogger(cfg.Mode)
	logger.Info("database connected")

	m := gormigrate.New(cfg.Db, gormigrate.DefaultOptions, migrations.All(cfg.Mode != gin.ReleaseMode))
	if err := m.Migrate(); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	gin.SetMode(cfg.Mode)
	router := gin.New()
	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		sentry.CurrentHub().Recover(err)
		logger.Error("panic recovered", "panic", fmt.Sprintf("%v", err))
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.Use(middleware.RequestLogger(logger))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     splitCSV(cfg.CORSAllowedOrigins),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	routes.RegisterPublicRoutes(router, cfg.Db)
	if cfg.Mode != gin.ReleaseMode {
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
	}

	userRepo := user.NewUsersRepository(cfg.Db)

	api := router.Group("/api", middleware.AuthMiddleware(userRepo))
	admin := api.Group("/admin", middleware.AdminMiddleware())

	routes.RegisterRoutes(api, cfg.Db)
	routes.RegisterAdminRoutes(admin, cfg.Db)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", "port", cfg.Port, "mode", cfg.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	logger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func initSentry() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return
	}
	_ = sentry.Init(sentry.ClientOptions{Dsn: dsn})
}

func initTimezone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	time.Local = ict
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
