package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISoundUseCases = (*SoundUseCases)(nil)

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
	uc.soundEventsRepository.Add(
		types.SoundEntity{SoundID: soundID, Loop: loop},
	)
}

func (uc *SoundUseCases) GetEvents() []types.SoundEntity {
	return uc.soundEventsRepository.Drain()
}

// StopAll останавливает все проигрываемые звуки через звуковой адаптер
func (uc *SoundUseCases) StopAll(
	soundPlayerAdapter interfaces.ISoundPlayerAdapter,
) {
	if soundPlayerAdapter != nil {
		soundPlayerAdapter.StopAll()
	}
	// Очищаем очередь событий, чтобы не проигрывать запланированные звуки
	uc.soundEventsRepository.Clear()
}
