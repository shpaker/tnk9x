package types

type EnemyAIDecision struct {
	Direction Direction
}

// Фазы поведения вражеского AI по времени этапа, как в оригинале:
// свободный обход -> охота на игрока -> атака штаба
const (
	EnemyAIPhaseWander = 1
	EnemyAIPhaseHunt   = 2
	EnemyAIPhaseSiege  = 3
)

// EnemyAIContext — контекст решения AI: текущая фаза этапа и цель,
// к которой смещается случайный обход
type EnemyAIContext struct {
	Phase     int
	TargetX   float64
	TargetY   float64
	HasTarget bool
}
