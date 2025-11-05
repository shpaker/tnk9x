package game

import (
	"github.com/shpaker/gonflict/internal/interfaces"
)

// GameRepositoriesRegistry содержит все игровые репозитории
type GameRepositoriesRegistry struct {
	bullets    interfaces.IBulletsRepository
	animations interfaces.IAnimationsRepository
	tanks      interfaces.ITanksRepository
}

// NewGameRepositoriesRegistry создает новый GameRepositoriesRegistry со всеми репозиториями
func NewGameRepositoriesRegistry() *GameRepositoriesRegistry {
	return &GameRepositoriesRegistry{
		bullets:    NewBulletsRepository(),
		animations: NewAnimationsRepository(),
		tanks:      NewTanksRepository(),
	}
}

// === Методы для доступа к репозиториям ===

// BulletsRepository возвращает репозиторий пуль
func (gr *GameRepositoriesRegistry) BulletsRepository() interfaces.IBulletsRepository {
	return gr.bullets
}

// AnimationsRepository возвращает репозиторий анимаций
func (gr *GameRepositoriesRegistry) AnimationsRepository() interfaces.IAnimationsRepository {
	return gr.animations
}

// TanksRepository возвращает репозиторий танков
func (gr *GameRepositoriesRegistry) TanksRepository() interfaces.ITanksRepository {
	return gr.tanks
}
