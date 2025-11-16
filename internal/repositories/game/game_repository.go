package game

import (
	"github.com/shpaker/tnk25/internal/interfaces"
)

type GameRepositoriesRegistry struct {
	bulletsRepository    interfaces.IBulletsRepository
	animationsRepository interfaces.IAnimationsRepository
	tanksRepository      interfaces.ITanksRepository
	bonusesRepository    interfaces.IBonusesRepository
}

func NewGameRepositoriesRegistry() *GameRepositoriesRegistry {
	return &GameRepositoriesRegistry{
		bulletsRepository:    NewBulletsRepository(),
		animationsRepository: NewAnimationsRepository(),
		tanksRepository:      NewTanksRepository(),
		bonusesRepository:    NewBonusesRepository(),
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

func (gr *GameRepositoriesRegistry) GetBonusesRepository() interfaces.IBonusesRepository {
	return gr.bonusesRepository
}
