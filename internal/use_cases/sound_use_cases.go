package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISoundUseCases = (*SoundUseCases)(nil)

// SoundUseCases — единственный канал управления звуком: и запуск,
// и остановка идут событиями через очередь кадра
type SoundUseCases struct {
	soundEventsRepository interfaces.ISoundEventsRepository
}

func NewSoundUseCases(
	soundEventsRepository interfaces.ISoundEventsRepository,
) *SoundUseCases {
	return &SoundUseCases{
		soundEventsRepository: soundEventsRepository,
	}
}

func (uc *SoundUseCases) RequestSound(soundID types.SoundID, loop bool) {
	action := types.SoundActionPlay
	if loop {
		action = types.SoundActionPlayLoop
	}
	uc.soundEventsRepository.Add(
		types.SoundEntity{Action: action, SoundID: soundID},
	)
}

func (uc *SoundUseCases) RequestStop(soundID types.SoundID) {
	uc.soundEventsRepository.Add(
		types.SoundEntity{Action: types.SoundActionStop, SoundID: soundID},
	)
}

func (uc *SoundUseCases) RequestStopAll() {
	uc.soundEventsRepository.Add(
		types.SoundEntity{Action: types.SoundActionStopAll},
	)
}

func (uc *SoundUseCases) GetEvents() []types.SoundEntity {
	return uc.soundEventsRepository.Drain()
}
