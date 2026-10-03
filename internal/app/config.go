package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x"
	"github.com/shpaker/tnk9x/internal/repositories/processed"
	"github.com/shpaker/tnk9x/internal/types"
)

type configSchema struct {
	App  appConfigSchema  `yaml:"app"`
	Game gameConfigSchema `yaml:"game"`
	Shop shopConfigSchema `yaml:"shop"`
}

// shopConfigSchema — товары магазина по порядку показа
type shopConfigSchema struct {
	Products []productConfigSchema `yaml:"products"`
}

// productConfigSchema — товар: ID в каталоге площадки, вид, пачка
// (для пачки) и число жетонов (для жетонов)
type productConfigSchema struct {
	ID     string `yaml:"id"`
	Kind   string `yaml:"kind"`
	Pack   int    `yaml:"pack"`
	Amount uint   `yaml:"amount"`
}

// productKinds — виды товаров в конфигурации
var productKinds = map[string]types.ProductKind{
	"no_ads":     types.ProductKindNoAds,
	"all_levels": types.ProductKindAllLevels,
	"pack":       types.ProductKindPack,
	"tokens":     types.ProductKindTokens,
}

type appConfigSchema struct {
	ScreenPx         [2]uint `yaml:"screen_px"`
	TitleFontSize    uint    `yaml:"title_font_size"`
	SubtitleFontSize uint    `yaml:"subtitle_font_size"`
	RegularFontSize  uint    `yaml:"regular_font_size"`
	GameTitle        string  `yaml:"game_title"`

	Languages         []string          `yaml:"languages"`
	DefaultLanguage   string            `yaml:"default_language"`
	LanguageFallbacks map[string]string `yaml:"language_fallbacks"`

	RatingAfterVictories uint `yaml:"rating_after_victories"`
}

type gameConfigSchema struct {
	EnemySpawners          [][2]int `yaml:"enemy_spawners"`
	Player1Spawn           [2]int   `yaml:"players_1_spawn_at"`
	Player2Spawn           [2]int   `yaml:"players_2_spawn_at"`
	HQPosition             [2]int   `yaml:"hq_position"`               // Позиция базы [x, y]
	EnemyRespawnDelayTicks uint     `yaml:"enemy_respawn_delay_ticks"` // Задержка между спавнами врагов в тиках

	BaseSizePx     uint   `yaml:"base_size_px"`
	MapBlocksCount [2]int `yaml:"map_blocks_count"` // Размер карты в блоках [width, height]

	SpawnPlayerSafeRadius float64                  `yaml:"spawn_player_safe_radius"`
	Campaign              string                   `yaml:"campaign"`
	DefaultLevel          defaultLevelConfigSchema `yaml:"default_level"`

	ShotCooldown     *bool `yaml:"shot_cooldown"`      // Перезарядка между выстрелами
	EnemyBonusPickup *bool `yaml:"enemy_bonus_pickup"` // Враги подбирают бонусы
}

// defaultLevelConfigSchema — дефолты уровня для карт без секций
type defaultLevelConfigSchema struct {
	MaxActive uint     `yaml:"max_active"`
	Time3Star uint     `yaml:"time_3star"`
	Waves     []string `yaml:"waves"`
}

type Config struct {
	ScreenPx         types.Size
	TitleFontSize    uint
	SubtitleFontSize uint
	RegularFontSize  uint
	GameTitle        string

	Languages types.LanguageConfig

	// RatingAfterVictories — с какой победы за сессию площадка может
	// попросить оценить игру; 0 — не просить
	RatingAfterVictories uint

	EnemySpawners          []types.Position
	Player1Spawn           types.Position
	Player2Spawn           types.Position
	HQPosition             [2]int
	EnemyRespawnDelayTicks uint

	BaseSizePx     uint
	MapBlocksCount types.Size

	TileBaseSize uint

	SpawnPlayerSafeRadius float64
	Campaign              string
	LevelDefaults         types.LevelDefaults

	ShotCooldown     bool
	EnemyBonusPickup bool

	// Products — товары магазина; пачки сверяются с кампанией
	// при её загрузке
	Products []types.ProductSpec
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
		TitleFontSize:        schema.App.TitleFontSize,
		SubtitleFontSize:     schema.App.SubtitleFontSize,
		RegularFontSize:      schema.App.RegularFontSize,
		GameTitle:            schema.App.GameTitle,
		RatingAfterVictories: schema.App.RatingAfterVictories,
		ShotCooldown:         true, // Значение по умолчанию
		EnemyBonusPickup:     true, // Значение по умолчанию
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

	languages, err := parseLanguageConfig(schema.App)
	if err != nil {
		return nil, err
	}
	cfg.Languages = languages

	if cfg.EnemyRespawnDelayTicks == 0 {
		cfg.EnemyRespawnDelayTicks = 2 * 60
	}

	cfg.SpawnPlayerSafeRadius = schema.Game.SpawnPlayerSafeRadius
	if cfg.SpawnPlayerSafeRadius <= 0 {
		cfg.SpawnPlayerSafeRadius = 4
	}

	cfg.Campaign = schema.Game.Campaign
	if cfg.Campaign == "" {
		cfg.Campaign = "main"
	}

	levelDefaults, err := processed.ParseLevelDefaults(
		schema.Game.DefaultLevel.MaxActive,
		schema.Game.DefaultLevel.Time3Star,
		cfg.EnemyRespawnDelayTicks,
		schema.Game.DefaultLevel.Waves,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid default_level in config: %w", err)
	}
	cfg.LevelDefaults = levelDefaults

	if schema.Game.ShotCooldown != nil {
		cfg.ShotCooldown = *schema.Game.ShotCooldown
	}

	if schema.Game.EnemyBonusPickup != nil {
		cfg.EnemyBonusPickup = *schema.Game.EnemyBonusPickup
	}

	products, err := parseProducts(schema.Shop)
	if err != nil {
		return nil, fmt.Errorf("invalid shop in config: %w", err)
	}
	cfg.Products = products

	return cfg, nil
}

// parseProducts — товары магазина: ID уникальны, вид известен,
// у пачки есть номер, у жетонов — количество
func parseProducts(schema shopConfigSchema) ([]types.ProductSpec, error) {
	products := make([]types.ProductSpec, 0, len(schema.Products))
	seen := make(map[string]bool)
	for _, product := range schema.Products {
		kind, ok := productKinds[product.Kind]
		switch {
		case product.ID == "":
			return nil, fmt.Errorf("product without id")
		case seen[product.ID]:
			return nil, fmt.Errorf("duplicate product %q", product.ID)
		case !ok:
			return nil, fmt.Errorf(
				"product %q: unknown kind %q", product.ID, product.Kind,
			)
		case kind == types.ProductKindPack && product.Pack < 1:
			return nil, fmt.Errorf("product %q: pack required", product.ID)
		case kind == types.ProductKindTokens && product.Amount == 0:
			return nil, fmt.Errorf("product %q: amount required", product.ID)
		}
		seen[product.ID] = true
		products = append(products, types.ProductSpec{
			ID:     product.ID,
			Kind:   kind,
			Pack:   product.Pack,
			Amount: product.Amount,
		})
	}
	return products, nil
}

// validateProductPacks — пачки товаров есть в кампании
func validateProductPacks(
	products []types.ProductSpec,
	campaign *types.CampaignEntity,
) error {
	for _, product := range products {
		if product.Kind != types.ProductKindPack {
			continue
		}
		if product.Pack > len(campaign.GetPacks()) {
			return fmt.Errorf(
				"product %q: no pack %d in campaign %s",
				product.ID, product.Pack, campaign.GetName(),
			)
		}
	}
	return nil
}

// parseLanguageConfig — языки интерфейса: без списка — только
// английский; язык по умолчанию — первый в списке, если не задан
func parseLanguageConfig(schema appConfigSchema) (types.LanguageConfig, error) {
	config := types.LanguageConfig{
		Default:   types.Language(schema.DefaultLanguage),
		Fallbacks: make(map[types.Language]types.Language),
	}
	for _, language := range schema.Languages {
		config.Languages = append(config.Languages, types.Language(language))
	}
	if len(config.Languages) == 0 {
		config.Languages = []types.Language{"en"}
	}
	if config.Default == types.LanguageAuto {
		config.Default = config.Languages[0]
	}
	if !config.IsSupported(config.Default) {
		return config, fmt.Errorf(
			"default_language %q is not in languages", config.Default,
		)
	}
	for from, to := range schema.LanguageFallbacks {
		if !config.IsSupported(types.Language(to)) {
			return config, fmt.Errorf(
				"language_fallbacks %s: %q is not in languages", from, to,
			)
		}
		config.Fallbacks[types.Language(from)] = types.Language(to)
	}
	return config, nil
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
