package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type APIConfig struct {
	BaseURL   string
	APIKey    string
	RateLimit int  // Requests per second
	BatchSize int  // Number of records per request
	UseAPI    bool // Whether to use API instead of direct DB
}

type AppConfig struct {
	DB        DatabaseConfig
	API       APIConfig
	BatchSize int
}

func LoadConfig() (*AppConfig, error) {
	// Load .env file if it exists
	godotenv.Load()

	return &AppConfig{
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "your-password"),
			DBName:   getEnv("DB_NAME", "your-database"),
		},
		API: APIConfig{
			BaseURL:   getEnv("API_BASE_URL", ""),
			APIKey:    getEnv("API_KEY", ""),
			RateLimit: getEnvAsInt("API_RATE_LIMIT", 60),
			BatchSize: getEnvAsInt("API_BATCH_SIZE", 100),
			UseAPI:    getEnvAsBool("USE_API", false),
		},
		BatchSize: getEnvAsInt("BATCH_SIZE", 1000),
	}, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return fallback
}
