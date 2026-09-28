package game_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/types"
)

// Очередь событий отдаётся один раз
func TestVisualEffectsRepository_DrainEvents(t *testing.T) {
	repository := game.NewVisualEffectsRepository()
	repository.AddEvent(types.VisualEventEntity{Kind: types.VisualEventShot})

	if got := len(repository.DrainEvents()); got != 1 {
		t.Fatalf("событий %d, ожидалось 1", got)
	}
	if got := len(repository.DrainEvents()); got != 0 {
		t.Errorf("после выборки событий %d, ожидалось 0", got)
	}
}

// Сверх лимита частицы и вспышки отбрасываются
func TestVisualEffectsRepository_Limits(t *testing.T) {
	repository := game.NewVisualEffectsRepository()
	for range 1000 {
		repository.AddParticle(types.ParticleEntity{Life: 1})
		repository.AddFlash(types.FlashEntity{Life: 1})
	}

	if got := len(repository.GetParticles()); got >= 1000 {
		t.Errorf("частиц %d: лимит не сработал", got)
	}
	if got := len(repository.GetFlashes()); got >= 1000 {
		t.Errorf("вспышек %d: лимит не сработал", got)
	}

	repository.SetParticles(nil)
	repository.SetFlashes(nil)
	if len(repository.GetParticles()) != 0 ||
		len(repository.GetFlashes()) != 0 {
		t.Error("набор не заменён")
	}
	repository.GetScreenShake().AddTrauma(0.5)
	if repository.GetScreenShake().GetTrauma() != 0.5 {
		t.Error("тряска должна быть одной на уровень")
	}
}
