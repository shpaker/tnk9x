package use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func TestSoundUseCases_RequestSoundAppendsInOrder(t *testing.T) {
	uc := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())

	uc.RequestSound(types.SoundIDFire, false)
	uc.RequestSound(types.SoundIDEngine, true)

	events := uc.GetEvents()
	if len(events) != 2 {
		t.Fatalf("ожидалось 2 события, получено %d", len(events))
	}
	if events[0].SoundID != types.SoundIDFire ||
		events[0].Action != types.SoundActionPlay {
		t.Errorf("событие 0: %+v", events[0])
	}
	if events[1].SoundID != types.SoundIDEngine ||
		events[1].Action != types.SoundActionPlayLoop {
		t.Errorf("событие 1: %+v", events[1])
	}
}

// GetEvents отдаёт события разрушающе: повторный вызов пуст
func TestSoundUseCases_GetEventsDrainsQueue(t *testing.T) {
	uc := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())
	uc.RequestSound(types.SoundIDExplosion, false)

	if got := len(uc.GetEvents()); got != 1 {
		t.Fatalf("первый вызов: ожидалось 1 событие, получено %d", got)
	}
	if got := len(uc.GetEvents()); got != 0 {
		t.Fatalf("второй вызов: ожидалось 0 событий, получено %d", got)
	}
}

// Остановка идёт тем же каналом, что и запуск: событием в очереди
func TestSoundUseCases_RequestStopProducesStopEvent(t *testing.T) {
	uc := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())

	uc.RequestSound(types.SoundIDEngine, true)
	uc.RequestStop(types.SoundIDEngine)

	events := uc.GetEvents()
	if len(events) != 2 {
		t.Fatalf("ожидалось 2 события, получено %d", len(events))
	}
	if events[1].SoundID != types.SoundIDEngine ||
		events[1].Action != types.SoundActionStop {
		t.Errorf("событие 1: %+v", events[1])
	}
}

func TestSoundUseCases_RequestStopAllProducesStopAllEvent(t *testing.T) {
	uc := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())

	uc.RequestSound(types.SoundIDBrick, false)
	uc.RequestStopAll()

	events := uc.GetEvents()
	if len(events) != 2 {
		t.Fatalf("ожидалось 2 события, получено %d", len(events))
	}
	if events[0].Action != types.SoundActionPlay {
		t.Errorf("событие 0: %+v", events[0])
	}
	if events[1].Action != types.SoundActionStopAll {
		t.Errorf("событие 1: %+v", events[1])
	}
}
