package config

import (
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host string
	User string
	Pass string
	Name string
}

type Config struct {
	Database DatabaseConfig
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host: getEnv("DBHOST", ""),
			User: getEnv("DBUSER", ""),
			Pass: getEnv("DBPASS", ""),
			Name: getEnv("DBNAME", ""),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}
