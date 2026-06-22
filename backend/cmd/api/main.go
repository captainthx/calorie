package main

import (
	"fmt"
	"log"
	"time"

	"github.com/captainthx/calorie/backend/internal/config"
	"github.com/captainthx/calorie/backend/internal/user"
	"github.com/gin-gonic/gin"
)

func main() {

	initTimezone()
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal("error in loading config: ", err)
	}

	createUserRoleEnumQuery :=
		`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_enum') THEN
			CREATE TYPE user_role_enum AS ENUM ('ADMIN', 'USER');
		END IF;
	END $$`

	if err := cfg.Db.Exec(createUserRoleEnumQuery).Error; err != nil {
		panic(fmt.Sprintf("Failed to create enum type: %v", err))
	}

	// AutoMigrate
	cfg.Db.AutoMigrate(&user.Users{})

	// Seed initial data
	var count int64
	cfg.Db.Model(&user.Users{}).Count(&count)
	if count == 0 {
		cfg.Db.Create(&[]user.Users{
			{Name: "John", Role: user.User, Token: "user-token-123", DaylyCalorieLimit: 2100, MounthlyPriceLimit: 10000},
			{Name: "Jane", Role: user.User, Token: "user-token-456", DaylyCalorieLimit: 2100, MounthlyPriceLimit: 10000},
			{Name: "Admin", Role: user.Admin, Token: "admin-token-789"},
		})
	}

	gin.SetMode(cfg.Mode)
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run(":" + cfg.Port)
}

func initTimezone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	time.Local = ict
}
