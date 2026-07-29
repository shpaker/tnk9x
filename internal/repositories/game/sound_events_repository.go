package game

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISoundEventsRepository = (*SoundEventsRepository)(nil)

// SoundEventsRepository — очередь звуковых событий текущего кадра
type SoundEventsRepository struct {
	events []types.SoundEntity
}

func NewSoundEventsRepository() *SoundEventsRepository {
	return &SoundEventsRepository{
		events: make([]types.SoundEntity, 0),
	}
}

func (r *SoundEventsRepository) Add(event types.SoundEntity) {
	r.events = append(r.events, event)
}

// Drain возвращает накопленные события и очищает очередь
func (r *SoundEventsRepository) Drain() []types.SoundEntity {
	events := r.events
	r.events = r.events[:0]
	return events
}

func (r *SoundEventsRepository) Clear() {
	r.events = r.events[:0]
}
