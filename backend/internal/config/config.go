package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Config struct {
	Port               string
	Mode               string
	CORSAllowedOrigins string
	Db                 *gorm.DB
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	mode := getEnvKey("MODE", "debug")
	port := getEnvKey("PORT", "8080")
	corsAllowedOrigins := getEnvKey("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	dbHost := getEnvKey("DB_HOST", "localhost")
	dbPort := getEnvKey("DB_PORT", "5432")
	dbUser := getEnvKey("DB_USERNAME", "myuser")
	dbPassword := getEnvKey("DB_PASSWORD", "mysecretpassword")
	dbName := getEnvKey("DB_NAME", "mydatabase")

	db, err := loadDb(dbPort, dbHost, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatal("Failed to connect to the Database")
		return nil, err
	}

	return &Config{
		Port:               port,
		Mode:               mode,
		CORSAllowedOrigins: corsAllowedOrigins,
		Db:                 db,
	}, nil
}

func getEnvKey(key string, defaultValue string) string {
	if val, exits := os.LookupEnv(key); exits {
		return val
	}
	return defaultValue
}
