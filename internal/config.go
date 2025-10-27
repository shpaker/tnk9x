package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig содержит настройки приложения
type AppConfig struct {
	Name string
}

// GameConfig содержит настройки игры
type GameConfig struct {
	EnemySpawners         [][]int `yaml:"enemy_spawners"`
	PlayerSpawners        [][]int `yaml:"players_spawners"`
	AIUpdateIntervalTicks int     `yaml:"ai_update_interval_ticks"` // Интервал обновления AI в тиках (по умолчанию 60 тиков = 1000мс)
}

// ConfigSchema структура для парсинга YAML файла
type ConfigSchema struct {
	App  AppConfig  `yaml:"app"`
	Game GameConfig `yaml:"game"`
}

// Config содержит конфигурацию приложения
type Config struct {
	AppConfig
	GameConfig
}

// ScreenWidth возвращает ширину экрана
func (c *Config) ScreenWidth() int {
	const (
		TileMinSize           = 8
		MapBlocksLength       = 26
		UpDownLeftPanelLength = 2
		RightPanelLength      = 4
	)
	MapWidthHeight := MapBlocksLength * TileMinSize
	return MapWidthHeight + UpDownLeftPanelLength*TileMinSize + RightPanelLength*TileMinSize
}

// ScreenHeight возвращает высоту экрана
func (c *Config) ScreenHeight() int {
	const (
		TileMinSize           = 8
		MapBlocksLength       = 26
		UpDownLeftPanelLength = 2
	)
	MapWidthHeight := MapBlocksLength * TileMinSize
	return MapWidthHeight + UpDownLeftPanelLength*TileMinSize*2
}

// GameSpeed возвращает скорость игры
func (c *Config) GameSpeed() float64 {
	return 1.0 / 60.0
}

// TileSize возвращает размер тайла
func (c *Config) TileSize() int {
	return 8
}

// LoadConfig загружает конфигурацию из файла config.yml
func LoadConfig() (*Config, error) {
	// Определяем путь к файлу конфигурации
	configPath := "config.yml"

	// Проверяем существование файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Читаем файл конфигурации
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Парсим YAML
	var configSchema ConfigSchema
	if err := yaml.Unmarshal(data, &configSchema); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Создаем конфигурацию
	cfg := &Config{
		AppConfig: AppConfig{
			Name: configSchema.App.Name,
		},
		GameConfig: configSchema.Game,
	}

	return cfg, nil
}
