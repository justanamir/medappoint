package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env           string
	Port          int
	DBHost        string
	DBPort        int
	DBName        string
	DBUser        string
	DBPassword    string
	DBSSLMode     string
	JWTSecret     string
	JWTIssuer     string
	JWTTTLMinutes int
}

func FromEnv() Config {
	return Config{
		Env:           getenv("APP_ENV", "dev"),
		Port:          getenvInt("PORT", 8080),
		DBHost:        getenv("DB_HOST", "localhost"),
		DBPort:        getenvInt("DB_PORT", 5432),
		DBName:        getenv("DB_NAME", "medappoint"),
		DBUser:        getenv("DB_USER", "medapp"),
		DBPassword:    getenv("DB_PASSWORD", "medpass"),
		DBSSLMode:     getenv("DB_SSLMODE", "disable"),
		JWTSecret:     getenv("JWT_SECRET", "dev-secret"),
		JWTIssuer:     getenv("JWT_ISSUER", "medappoint"),
		JWTTTLMinutes: getenvInt("JWT_TTL_MINUTES", 60),
	}
}

func (c Config) PGConnString(hostOverride string) string {
	host := c.DBHost
	if hostOverride != "" {
		host = hostOverride
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, host, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
