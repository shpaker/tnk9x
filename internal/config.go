package internal

import (
	"os"
	"strconv"
)

// AppConfig содержит настройки приложения
type AppConfig struct {
	Name         string  `json:"name"`
	ScreenWidth  int     `json:"screen_width"`
	ScreenHeight int     `json:"screen_height"`
	GameSpeed    float64 `json:"game_speed"`
	TileSize     int     `json:"tile_size"`
}

// Config содержит конфигурацию приложения
type Config struct {
	AppConfig
}

// Load загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	// Константы для вычисления размеров экрана
	const (
		TileMinSize           = 8
		MapBlocksLength       = 26
		UpDownLeftPanelLength = 2
		RightPanelLength      = 4
		DefaultGameSpeed      = 1.0 / 60.0
	)

	MapWidthHeight := MapBlocksLength * TileMinSize
	ScreenWidth := MapWidthHeight + UpDownLeftPanelLength*TileMinSize + RightPanelLength*TileMinSize
	ScreenHeight := MapWidthHeight + UpDownLeftPanelLength*TileMinSize*2

	cfg := &Config{
		AppConfig{
			Name:         getEnv("APP_NAME", "gonflict"),
			ScreenWidth:  getEnvAsInt("SCREEN_WIDTH", ScreenWidth),
			ScreenHeight: getEnvAsInt("SCREEN_HEIGHT", ScreenHeight),
			GameSpeed:    getEnvAsFloat64("GAME_SPEED", DefaultGameSpeed),
			TileSize:     getEnvAsInt("TILE_SIZE", TileMinSize),
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

// getEnvAsFloat64 получает переменную окружения как float64 или возвращает значение по умолчанию
func getEnvAsFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
