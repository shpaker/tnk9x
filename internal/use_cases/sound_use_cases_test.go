package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func TestSoundUseCases_RequestSoundAppendsInOrder(t *testing.T) {
	uc := use_cases.NewSoundUseCases()

	uc.RequestSound(types.SoundIDFire, false)
	uc.RequestSound(types.SoundIDEngine, true)

	events := uc.GetEvents()
	if len(events) != 2 {
		t.Fatalf("ожидалось 2 события, получено %d", len(events))
	}
	if events[0].SoundID != types.SoundIDFire || events[0].Loop {
		t.Errorf("событие 0: %+v", events[0])
	}
	if events[1].SoundID != types.SoundIDEngine || !events[1].Loop {
		t.Errorf("событие 1: %+v", events[1])
	}
}

// GetEvents отдаёт события разрушающе: повторный вызов пуст
func TestSoundUseCases_GetEventsDrainsQueue(t *testing.T) {
	uc := use_cases.NewSoundUseCases()
	uc.RequestSound(types.SoundIDExplosion, false)

	if got := len(uc.GetEvents()); got != 1 {
		t.Fatalf("первый вызов: ожидалось 1 событие, получено %d", got)
	}
	if got := len(uc.GetEvents()); got != 0 {
		t.Fatalf("второй вызов: ожидалось 0 событий, получено %d", got)
	}
}

func TestSoundUseCases_StopAll(t *testing.T) {
	uc := use_cases.NewSoundUseCases()
	player := &testutil.FakeSoundPlayer{}

	uc.RequestSound(types.SoundIDBrick, false)
	uc.RequestSound(types.SoundIDSteel, false)

	uc.StopAll(player)

	if player.StopAllCalls != 1 {
		t.Errorf(
			"ожидался 1 вызов StopAll адаптера, было %d",
			player.StopAllCalls,
		)
	}
	if got := len(uc.GetEvents()); got != 0 {
		t.Errorf("очередь не очищена: %d событий", got)
	}
}

func TestSoundUseCases_StopAllWithNilAdapter(t *testing.T) {
	uc := use_cases.NewSoundUseCases()
	uc.RequestSound(types.SoundIDBonus, false)

	uc.StopAll(nil) // не должно паниковать

	if got := len(uc.GetEvents()); got != 0 {
		t.Errorf("очередь не очищена: %d событий", got)
	}
}
