package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppPort string
	AppEnv  string
	DB      DatabaseConfig
	JWT     JWTConfig
	Log     LogConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type LogConfig struct {
	Dir        string
	Level      string
	MaxSize    int  // in megabytes
	MaxBackups int  // number of backup files
	MaxAge     int  // retention in days
	Compress   bool // gzip compression
}

func LoadConfig() *Config {
	// Load .env file if it exists, otherwise fall back to environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: .env file not found, loading configuration from system environment")
	}

	return &Config{
		AppName: getEnv("APP_NAME", "test-golang-project"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "test_db"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "default_secret_key"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		Log: LogConfig{
			Dir:        getEnv("LOG_DIR", "storage/logs"),
			Level:      getEnv("LOG_LEVEL", "info"),
			MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 10),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 14),
			MaxAge:     getEnvAsInt("LOG_MAX_AGE", 14),
			Compress:   getEnvAsBool("LOG_COMPRESS", true),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return fallback
}
