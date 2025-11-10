package game

import (
	"github.com/shpaker/gonflict/internal/interfaces"
)

// GameRepositoriesRegistry содержит все игровые репозитории.
type GameRepositoriesRegistry struct {
	// Repositories
	bulletsRepository    interfaces.IBulletsRepository
	animationsRepository interfaces.IAnimationsRepository
	tanksRepository      interfaces.ITanksRepository
}

// NewGameRepositoriesRegistry создает новый GameRepositoriesRegistry со всеми репозиториями
func NewGameRepositoriesRegistry() *GameRepositoriesRegistry {
	return &GameRepositoriesRegistry{
		bulletsRepository:    NewBulletsRepository(),
		animationsRepository: NewAnimationsRepository(),
		tanksRepository:      NewTanksRepository(),
	}
}

// === Методы для доступа к репозиториям ===

// GetBulletsRepository возвращает репозиторий пуль.
func (gr *GameRepositoriesRegistry) GetBulletsRepository() interfaces.IBulletsRepository {
	return gr.bulletsRepository
}

// GetAnimationsRepository возвращает репозиторий анимаций.
func (gr *GameRepositoriesRegistry) GetAnimationsRepository() interfaces.IAnimationsRepository {
	return gr.animationsRepository
}

// GetTanksRepository возвращает репозиторий танков.
func (gr *GameRepositoriesRegistry) GetTanksRepository() interfaces.ITanksRepository {
	return gr.tanksRepository
}
