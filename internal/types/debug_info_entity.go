package types

// DebugInfoData хранит сведения для отладки игры
type DebugInfoData struct {
	FPS                float64
	TPS                float64
	PlayerLives        uint
	PlayerInitialLives uint
	TotalEnemies       uint
	RemainingEnemies   uint
}
