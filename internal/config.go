package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/gonflict/internal/types"
)

// configSchema структура для парсинга YAML файла
type configSchema struct {
	App  appConfigSchema  `yaml:"app"`
	Game gameConfigSchema `yaml:"game"`
}

// appConfigSchema для парсинга app секции из YAML
type appConfigSchema struct {
	Name        string  `yaml:"name"`
	LevelNumber int     `yaml:"level_number"`
	ScreenPx    [2]uint `yaml:"screen_px"`
}

// gameConfigSchema для парсинга game секции из YAML
type gameConfigSchema struct {
	EnemySpawners         [][2]int `yaml:"enemy_spawners"`
	PlayerSpawners        [][2]int `yaml:"players_spawners"`
	HQPosition            [2]int   `yaml:"hq_position"`              // Позиция базы [x, y]
	AIUpdateIntervalTicks int      `yaml:"ai_update_interval_ticks"` // Интервал обновления AI в тиках (по умолчанию 60 тиков = 1000мс)

	// Игровые константы
	BaseSizePx     uint    `yaml:"base_size_px"`
	MapBlocksCount [2]int  `yaml:"map_blocks_count"` // Размер карты в блоках [width, height]
	MapOffsets     [2]uint `yaml:"map_offsets_px"`   // Оффсеты игровой карты от угла экрана [x, y]
}

// Config содержит конфигурацию приложения
type Config struct {
	// App settings
	Name        string
	LevelNumber int
	ScreenPx    types.Size

	// Game settings
	EnemySpawners         [][2]int
	PlayerSpawners        [][2]int
	HQPosition            [2]int
	AIUpdateIntervalTicks int

	// Game constants
	BaseSizePx     uint
	MapBlocksCount types.Size
	MapOffsets     [2]uint

	// Вычисляемые значения
	TileBaseSize uint // Вычисляется как base_size_px / 2
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
	var schema configSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Создаем конфигурацию
	cfg := &Config{
		Name:        schema.App.Name,
		LevelNumber: schema.App.LevelNumber,
		ScreenPx: types.Size{
			Width:  int(schema.App.ScreenPx[0]),
			Height: int(schema.App.ScreenPx[1]),
		},

		EnemySpawners:         schema.Game.EnemySpawners,
		PlayerSpawners:        schema.Game.PlayerSpawners,
		HQPosition:            schema.Game.HQPosition,
		AIUpdateIntervalTicks: schema.Game.AIUpdateIntervalTicks,

		BaseSizePx: schema.Game.BaseSizePx,
		MapBlocksCount: types.Size{
			Width:  schema.Game.MapBlocksCount[0],
			Height: schema.Game.MapBlocksCount[1],
		},
		MapOffsets: schema.Game.MapOffsets,
	}

	// Вычисляем TileBaseSize как base_size_px / 2
	cfg.TileBaseSize = cfg.BaseSizePx / 2

	return cfg, nil
}

// Геттеры для интерфейса IConfigProvider

func (c *Config) GetEnemySpawners() [][2]int {
	return c.EnemySpawners
}

func (c *Config) GetPlayerSpawners() [][2]int {
	return c.PlayerSpawners
}

func (c *Config) GetHQPosition() [2]int {
	return c.HQPosition
}

func (c *Config) GetAIUpdateIntervalTicks() int {
	return c.AIUpdateIntervalTicks
}

func (c *Config) GetBaseSizePx() uint {
	return c.BaseSizePx
}

func (c *Config) GetMapBlocksCount() types.Size {
	return c.MapBlocksCount
}

func (c *Config) GetMapOffsets() [2]uint {
	return c.MapOffsets
}

func (c *Config) GetTileBaseSize() uint {
	return c.TileBaseSize
}

// ScreenWidth возвращает ширину экрана
func (c *Config) ScreenWidth() int {
	return c.ScreenPx.Width
}

// ScreenHeight возвращает высоту экрана
func (c *Config) ScreenHeight() int {
	return c.ScreenPx.Height
}

// GameSpeed возвращает скорость игры
func (c *Config) GameSpeed() float64 {
	return 1.0 / 60.0
}

// TileSize возвращает размер тайла
func (c *Config) TileSize() int {
	return int(c.TileBaseSize)
}
