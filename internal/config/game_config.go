package config

// GameConfig содержит настройки игры
type GameConfig struct {
	LevelNumber           int      `yaml:"level_number"` // Номер уровня для загрузки
	EnemySpawners         [][2]int `yaml:"enemy_spawners"`
	PlayerSpawners        [][2]int `yaml:"players_spawners"`
	AIUpdateIntervalTicks int      `yaml:"ai_update_interval_ticks"` // Интервал обновления AI в тиках (по умолчанию 60 тиков = 1000мс)
}
