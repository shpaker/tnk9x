package states

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

// fakeStageUseCases — только флаг паузы; остальные методы
// в сценарии меню паузы не вызываются
type fakeStageUseCases struct {
	interfaces.IStageUseCases
	paused bool
}

func (f *fakeStageUseCases) TogglePause()      { f.paused = !f.paused }
func (f *fakeStageUseCases) IsPaused() bool    { return f.paused }
func (f *fakeStageUseCases) ResumeStageState() { f.paused = false }

// fakeSoundUseCases записывает запрошенные остановки звуков
type fakeSoundUseCases struct {
	stopped []types.SoundID
}

var _ interfaces.ISoundUseCases = (*fakeSoundUseCases)(nil)

func (f *fakeSoundUseCases) RequestSound(types.SoundID, bool) {}

func (f *fakeSoundUseCases) RequestStop(soundID types.SoundID) {
	f.stopped = append(f.stopped, soundID)
}

func (f *fakeSoundUseCases) RequestStopAll() {}

func (f *fakeSoundUseCases) GetEvents() []types.SoundEntity { return nil }

func TestStageState_PauseStopsEngine(t *testing.T) {
	stageUseCases := &fakeStageUseCases{}
	soundUseCases := &fakeSoundUseCases{}
	input := &testutil.FakeMenuInput{}
	state := NewStageState(StageStateDependencies{
		StageUseCases: stageUseCases,
		SoundUseCases: soundUseCases,
		MenuInput:     input,
	})
	frame := func(pressed testutil.FakeMenuInput) {
		*input = pressed
		state.handlePauseMenu()
		input.Reset()
	}
	engineStops := func() int {
		count := 0
		for _, soundID := range soundUseCases.stopped {
			if soundID == types.SoundIDEngine {
				count++
			}
		}
		return count
	}

	// Открытие меню глушит луп двигателя
	frame(testutil.FakeMenuInput{Pause: true})
	if !stageUseCases.paused || engineStops() != 1 {
		t.Fatalf("пауза глушит двигатель: пауза %v, остановок %d",
			stageUseCases.paused, engineStops())
	}

	// Пока меню открыто, остановка не повторяется
	frame(testutil.FakeMenuInput{})
	if engineStops() != 1 {
		t.Errorf("остановка двигателя — один раз на открытие, их %d", engineStops())
	}

	// CONTINUE снимает паузу; после кадра игры повторное открытие
	// снова глушит двигатель
	frame(testutil.FakeMenuInput{Confirm: true})
	if stageUseCases.paused {
		t.Fatal("CONTINUE снимает паузу")
	}
	frame(testutil.FakeMenuInput{})
	frame(testutil.FakeMenuInput{Pause: true})
	if engineStops() != 2 {
		t.Errorf("каждое открытие меню глушит двигатель, остановок %d", engineStops())
	}
}
