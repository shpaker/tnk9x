package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x"
	"github.com/shpaker/tnk9x/internal/types"
)

type configSchema struct {
	App  appConfigSchema  `yaml:"app"`
	Game gameConfigSchema `yaml:"game"`
}

type appConfigSchema struct {
	ScreenPx         [2]uint  `yaml:"screen_px"`
	TitleFontSize    uint     `yaml:"title_font_size"`
	SubtitleFontSize uint     `yaml:"subtitle_font_size"`
	RegularFontSize  uint     `yaml:"regular_font_size"`
	GameTitle        string   `yaml:"game_title"`
	Volume           *float64 `yaml:"volume"`
}

type gameConfigSchema struct {
	EnemySpawners          [][2]int `yaml:"enemy_spawners"`
	Player1Spawn           [2]int   `yaml:"players_1_spawn_at"`
	Player2Spawn           [2]int   `yaml:"players_2_spawn_at"`
	HQPosition             [2]int   `yaml:"hq_position"`               // Позиция базы [x, y]
	AIUpdateIntervalTicks  int      `yaml:"ai_update_interval_ticks"`  // Интервал обновления AI в тиках (по умолчанию 60 тиков = 1000мс)
	EnemyRespawnDelayTicks uint     `yaml:"enemy_respawn_delay_ticks"` // Задержка между спавнами врагов в тиках

	BaseSizePx     uint   `yaml:"base_size_px"`
	MapBlocksCount [2]int `yaml:"map_blocks_count"` // Размер карты в блоках [width, height]
}

type Config struct {
	ScreenPx         types.Size
	TitleFontSize    uint
	SubtitleFontSize uint
	RegularFontSize  uint
	GameTitle        string
	Volume           float64

	EnemySpawners          []types.Position
	Player1Spawn           types.Position
	Player2Spawn           types.Position
	HQPosition             [2]int
	AIUpdateIntervalTicks  int
	EnemyRespawnDelayTicks uint

	BaseSizePx     uint
	MapBlocksCount types.Size

	TileBaseSize uint
}

func LoadConfig() (*Config, error) {
	// Диск в приоритете (правки пользователя рядом с бинарником),
	// иначе — встроенная копия (wasm, самодостаточный бинарник)
	data, err := os.ReadFile("config.yml")
	if err != nil {
		data, err = tnk9x.FS.ReadFile("config.yml")
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var schema configSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg := &Config{
		ScreenPx: types.Size{
			Width:  int(schema.App.ScreenPx[0]),
			Height: int(schema.App.ScreenPx[1]),
		},
		TitleFontSize:    schema.App.TitleFontSize,
		SubtitleFontSize: schema.App.SubtitleFontSize,
		RegularFontSize:  schema.App.RegularFontSize,
		GameTitle:        schema.App.GameTitle,
		Volume:           1.0, // Значение по умолчанию
		EnemySpawners: convertCoordsToPositions(
			schema.Game.EnemySpawners,
		),
		Player1Spawn: convertCoordToPosition(
			schema.Game.Player1Spawn,
		),
		Player2Spawn: convertCoordToPosition(
			schema.Game.Player2Spawn,
		),
		HQPosition:             schema.Game.HQPosition,
		AIUpdateIntervalTicks:  schema.Game.AIUpdateIntervalTicks,
		EnemyRespawnDelayTicks: schema.Game.EnemyRespawnDelayTicks,

		BaseSizePx: schema.Game.BaseSizePx,
		MapBlocksCount: types.Size{
			Width:  schema.Game.MapBlocksCount[0],
			Height: schema.Game.MapBlocksCount[1],
		},
	}

	cfg.TileBaseSize = cfg.BaseSizePx / 2

	// Единая нормализация размеров шрифтов: дальше по графу зависимостей
	// значения считаются заданными и не проверяются
	if cfg.TitleFontSize == 0 {
		cfg.TitleFontSize = 32
	}
	if cfg.RegularFontSize == 0 {
		cfg.RegularFontSize = 8
	}
	if cfg.SubtitleFontSize == 0 {
		cfg.SubtitleFontSize = cfg.RegularFontSize
	}

	if cfg.EnemyRespawnDelayTicks == 0 {
		cfg.EnemyRespawnDelayTicks = 3 * 60
	}

	// Если громкость указана в конфиге, используем её значение
	if schema.App.Volume != nil {
		cfg.Volume = *schema.App.Volume
		// Валидация громкости: должна быть от 0.0 до 1.0
		if cfg.Volume < 0.0 {
			cfg.Volume = 0.0
		} else if cfg.Volume > 1.0 {
			cfg.Volume = 1.0
		}
	}

	return cfg, nil
}

func convertCoordsToPositions(coords [][2]int) []types.Position {
	positions := make([]types.Position, len(coords))
	for i, coord := range coords {
		positions[i] = convertCoordToPosition(coord)
	}
	return positions
}

func convertCoordToPosition(coord [2]int) types.Position {
	return types.Position{
		X: float64(coord[0]),
		Y: float64(coord[1]),
	}
}

func (c *Config) GetEnemySpawners() []types.Position {
	return c.EnemySpawners
}

func (c *Config) GetPlayer1Spawn() types.Position {
	return c.Player1Spawn
}

func (c *Config) GetPlayer2Spawn() types.Position {
	return c.Player2Spawn
}

func (c *Config) GetHQPosition() [2]int {
	return c.HQPosition
}

func (c *Config) GetAIUpdateIntervalTicks() int {
	return c.AIUpdateIntervalTicks
}

func (c *Config) GetEnemyRespawnDelayTicks() uint {
	return c.EnemyRespawnDelayTicks
}

func (c *Config) GetBaseSizePx() uint {
	return c.BaseSizePx
}

func (c *Config) GetMapBlocksCount() types.Size {
	return c.MapBlocksCount
}

func (c *Config) GetTileBaseSize() uint {
	return c.TileBaseSize
}

func (c *Config) GetTitleFontSize() uint {
	return c.TitleFontSize
}

func (c *Config) GetSubtitleFontSize() uint {
	return c.SubtitleFontSize
}

func (c *Config) GetRegularFontSize() uint {
	return c.RegularFontSize
}

func (c *Config) GetGameTitle() string {
	return c.GameTitle
}

func (c *Config) GetVolume() float64 {
	return c.Volume
}

func (c *Config) ScreenWidth() int {
	return c.ScreenPx.Width
}

func (c *Config) ScreenHeight() int {
	return c.ScreenPx.Height
}

func (c *Config) GameSpeed() float64 {
	return 1.0 / 60.0
}

func (c *Config) TileSize() int {
	return int(c.TileBaseSize)
}
