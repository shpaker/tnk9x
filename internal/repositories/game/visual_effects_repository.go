package game

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Пределы эффектов уровня: сверх них новые частицы и вспышки
// отбрасываются, чтобы серия взрывов не просаживала кадр
const (
	maxParticles = 400
	maxFlashes   = 8
)

var _ interfaces.IVisualEffectsRepository = (*VisualEffectsRepository)(nil)

// VisualEffectsRepository реализует IVisualEffectsRepository
type VisualEffectsRepository struct {
	events    []types.VisualEventEntity
	particles []types.ParticleEntity
	flashes   []types.FlashEntity
	shake     types.ScreenShakeEntity
}

func NewVisualEffectsRepository() *VisualEffectsRepository {
	return &VisualEffectsRepository{
		events:    make([]types.VisualEventEntity, 0),
		particles: make([]types.ParticleEntity, 0, maxParticles),
		flashes:   make([]types.FlashEntity, 0, maxFlashes),
	}
}

// События

func (r *VisualEffectsRepository) AddEvent(event types.VisualEventEntity) {
	r.events = append(r.events, event)
}

func (r *VisualEffectsRepository) DrainEvents() []types.VisualEventEntity {
	events := r.events
	r.events = r.events[:0]
	return events
}

// Частицы

func (r *VisualEffectsRepository) AddParticle(particle types.ParticleEntity) {
	if len(r.particles) >= maxParticles {
		return
	}
	r.particles = append(r.particles, particle)
}

func (r *VisualEffectsRepository) GetParticles() []types.ParticleEntity {
	return r.particles
}

func (r *VisualEffectsRepository) SetParticles(
	particles []types.ParticleEntity,
) {
	r.particles = particles
}

// Вспышки

func (r *VisualEffectsRepository) AddFlash(flash types.FlashEntity) {
	if len(r.flashes) >= maxFlashes {
		return
	}
	r.flashes = append(r.flashes, flash)
}

func (r *VisualEffectsRepository) GetFlashes() []types.FlashEntity {
	return r.flashes
}

func (r *VisualEffectsRepository) SetFlashes(flashes []types.FlashEntity) {
	r.flashes = flashes
}

// Тряска

func (r *VisualEffectsRepository) GetScreenShake() *types.ScreenShakeEntity {
	return &r.shake
}
