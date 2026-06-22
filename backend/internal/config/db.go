package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func loadDb(dbPort string, dbHost string, dbUser string, dbPassword string, dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("Failed to connect to the Database")
		return nil, err
	}

	err = sqlDb.Ping()
	if err != nil {
		log.Fatal("Failed to ping the Database")
		return nil, err
	}
	fmt.Println("🚀 Connected Successfully to the Database")
	return db, err
}
