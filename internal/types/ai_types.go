package types

// EnemyAIDecision решение AI для врага
type EnemyAIDecision struct {
	Direction Direction // Направление движения
}

// GameAiContext представляет контекст игры для AI
type GameAiContext struct {
	Player  *TankEntity    // Игрок
	Enemies []*TankEntity  // Враги
	Bullets []BulletEntity // Пули
	Blocks  []BlockEntity  // Блоки/стены
}
