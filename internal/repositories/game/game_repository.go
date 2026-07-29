package game

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IGameRepositoriesRegistry = (*GameRepositoriesRegistry)(nil)

type GameRepositoriesRegistry struct {
	bulletsRepository     interfaces.IBulletsRepository
	animationsRepository  interfaces.IAnimationsRepository
	tanksRepository       interfaces.ITanksRepository
	bonusesRepository     interfaces.IBonusesRepository
	soundEventsRepository interfaces.ISoundEventsRepository
}

func NewGameRepositoriesRegistry() *GameRepositoriesRegistry {
	return &GameRepositoriesRegistry{
		bulletsRepository:     NewBulletsRepository(),
		animationsRepository:  NewAnimationsRepository(),
		tanksRepository:       NewTanksRepository(),
		bonusesRepository:     NewBonusesRepository(),
		soundEventsRepository: NewSoundEventsRepository(),
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

func (gr *GameRepositoriesRegistry) GetSoundEventsRepository() interfaces.ISoundEventsRepository {
	return gr.soundEventsRepository
}
