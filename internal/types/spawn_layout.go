package types

// SpawnLayout — конфигурация точек спавна уровня.
type SpawnLayout struct {
	EnemySpawners  []Position
	Player1Spawner Position
	Player2Spawner Position
	BaseSize       Size
}
