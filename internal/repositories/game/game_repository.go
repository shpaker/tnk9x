package game

import (
	"github.com/shpaker/gonflict/internal/interfaces"
)

type GameRepositoriesRegistry struct {
	bulletsRepository    interfaces.IBulletsRepository
	animationsRepository interfaces.IAnimationsRepository
	tanksRepository      interfaces.ITanksRepository
}

func NewGameRepositoriesRegistry() *GameRepositoriesRegistry {
	return &GameRepositoriesRegistry{
		bulletsRepository:    NewBulletsRepository(),
		animationsRepository: NewAnimationsRepository(),
		tanksRepository:      NewTanksRepository(),
	}
}

func (gr *GameRepositoriesRegistry) GetBulletsRepository() interfaces.IBulletsRepository {
	return gr.bulletsRepository
}

func (gr *GameRepositoriesRegistry) GetAnimationsRepository() interfaces.IAnimationsRepository {
	return gr.animationsRepository
}

func (gr *GameRepositoriesRegistry) GetTanksRepository() interfaces.ITanksRepository {
	return gr.tanksRepository
}
