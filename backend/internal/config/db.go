package config

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func loadDb(dbPort string, dbHost string, dbUser string, dbPassword string, dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok", dbHost, dbUser, dbPassword, dbName, dbPort)
	return loadDbFromDSN(dsn)
}

func loadDbFromDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database handle: %w", err)
	}

	if err := sqlDb.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	sqlDb.SetMaxOpenConns(25)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
