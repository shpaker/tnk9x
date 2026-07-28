package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISoundUseCases = (*SoundUseCases)(nil)

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

// StopAll останавливает все проигрываемые звуки через звуковой адаптер
func (uc *SoundUseCases) StopAll(
	soundPlayerAdapter interfaces.ISoundPlayerAdapter,
) {
	if soundPlayerAdapter != nil {
		soundPlayerAdapter.StopAll()
	}
	// Очищаем очередь событий, чтобы не проигрывать запланированные звуки
	uc.events = uc.events[:0]
}
