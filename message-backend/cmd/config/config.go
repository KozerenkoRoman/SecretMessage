package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIPort    int
	SocketPort int
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBPoolMax  int
	DBPoolMin  int
	DBLogging  bool

	AdminUser  string
	AdminEmail string
	AdminPass  string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	conf := &Config{
		APIPort:    getEnvAsInt("PORT", 8080),
		SocketPort: getEnvAsInt("SOCKET_PORT", 3000),

		DBHost:     getEnv("POSTGRES_HOST", "127.0.0.1"),
		DBPort:     getEnvAsInt("POSTGRES_PORT", 5432),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "111"),
		DBName:     getEnv("POSTGRES_DB", "love_letter"),
		DBPoolMax:  getEnvAsInt("POSTGRES_POOL_MAX", 20),
		DBPoolMin:  getEnvAsInt("POSTGRES_POOL_MIN", 5),
		DBLogging:  getEnvAsBool("POSTGRES_LOGGING", false),

		AdminUser:  getEnv("ADMIN_USER", "admin"),
		AdminEmail: getEnv("ADMIN_EMAIL", "admin@loveletter.local"),
		AdminPass:  getEnv("ADMIN_PASSWORD", "SuperSecureAdminPassword2026"),
	}

	if conf.DBPassword == "" {
		panic("Config error: POSTGRES_PASSWORD is required and cannot be empty")
	}
	if conf.DBName == "" {
		panic("Config error: POSTGRES_DB name is required")
	}

	return conf
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return fallback
}
