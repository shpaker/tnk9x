package config

import (
	"os"
	"strconv"
)

// AppConfig содержит настройки приложения
type AppConfig struct {
	Name string `json:"name"`
}

// LogConfig содержит настройки логирования
type LogConfig struct {
	Level string `json:"level"`
}

// Config содержит конфигурацию приложения
type Config struct {
	AppConfig
	LogConfig
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		AppConfig{
			Name: getEnv("APP_NAME", "gonflict"),
		},
		LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	return cfg, nil
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает переменную окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
