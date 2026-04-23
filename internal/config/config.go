package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	ServerPort       string
	LoginMaxAttempts int
	LoginWindowMin   int
	LoginBlockMin    int
}

func Load() *Config {
	return &Config{
		DBHost:           getEnv("DB_HOST", "127.0.0.1"),
		DBPort:           getEnv("DB_PORT", "3306"),
		DBUser:           getEnv("DB_USER", "root"),
		DBPassword:       getEnv("DB_PASSWORD", "root"),
		DBName:           getEnv("DB_NAME", "flowmodel"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		LoginMaxAttempts: getEnvInt("LOGIN_MAX_ATTEMPTS", 5),
		LoginWindowMin:   getEnvInt("LOGIN_WINDOW_MINUTES", 15),
		LoginBlockMin:    getEnvInt("LOGIN_BLOCK_MINUTES", 15),
	}
}

func (c *Config) DSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
