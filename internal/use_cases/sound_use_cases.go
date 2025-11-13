package use_cases

import "github.com/shpaker/tnk25/internal/types"

type SoundUseCases struct {
	events []types.SoundEntity
}

func NewSoundUseCases() *SoundUseCases {
	return &SoundUseCases{
		events: make([]types.SoundEntity, 0),
	}
}

func (uc *SoundUseCases) RequestSound(soundID types.SoundID, loop bool) {
	uc.events = append(
		uc.events,
		types.SoundEntity{SoundID: soundID, Loop: loop},
	)
}

func (uc *SoundUseCases) GetEvents() []types.SoundEntity {
	events := uc.events
	uc.events = uc.events[:0] // Очищаем после получения
	return events
}
