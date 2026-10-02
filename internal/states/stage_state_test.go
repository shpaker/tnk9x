package states

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
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

// fakeFinishingStageUseCases — уровень завершается поражением
// во время обновления игровых объектов
type fakeFinishingStageUseCases struct {
	fakeStageUseCases
	finished bool
}

func (f *fakeFinishingStageUseCases) UpdateGameObjects(float64) { f.finished = true }
func (f *fakeFinishingStageUseCases) IsStageFinished() bool     { return f.finished }
func (f *fakeFinishingStageUseCases) IsStageWon() bool          { return false }
func (f *fakeFinishingStageUseCases) PauseStageState()          { f.paused = true }
func (f *fakeFinishingStageUseCases) CanRevivePlayers() bool    { return false }

func (f *fakeFinishingStageUseCases) TryRespawnPlayersTanks() (
	*types.TankEntity, *types.TankEntity,
) {
	return nil, nil
}

func (f *fakeFinishingStageUseCases) TrySpawnEnemy() *types.TankEntity {
	return nil
}

func (f *fakeFinishingStageUseCases) GetStageResult() types.StageResult {
	return types.StageResult{}
}

// fakeStageWorld — пустой мир уровня: танков, тайлов, бонусов,
// фар и AI нет
type fakeStageWorld struct {
	interfaces.ITankLifecycleUseCases
	interfaces.ITilesUseCases
	interfaces.ILightingUseCases
	interfaces.IAiInputAdapter
	interfaces.IBonusesRepository
	interfaces.IProgressionUseCases
}

func (fakeStageWorld) UpdateAllTanksLifecycle() error                       { return nil }
func (fakeStageWorld) UpdateAnimations()                                    {}
func (fakeStageWorld) UpdateHeadlights()                                    {}
func (fakeStageWorld) Update(float64)                                       {}
func (fakeStageWorld) GetAllBonuses() []*types.BonusEntity                  { return nil }
func (fakeStageWorld) CalcStars(types.StageResult, *types.LevelEntity) uint { return 0 }
func (fakeStageWorld) GetLevelStars(int) uint                               { return 0 }
func (fakeStageWorld) NextLevel(level int) (int, bool)                      { return level + 1, false }

// Итоги строятся в том же кадре, где уровень завершился: иначе Draw
// видит паузу без итогов и на кадр рисует меню паузы
func TestStageState_FinishBuildsResultSameFrame(t *testing.T) {
	stageUseCases := &fakeFinishingStageUseCases{}
	world := fakeStageWorld{}
	state := NewStageState(StageStateDependencies{
		TankCommonUseCases:    &testutil.FakeTankCommonUseCases{},
		TankLifecycleUseCases: world,
		TilesUseCases:         world,
		StageUseCases:         stageUseCases,
		SoundUseCases:         &fakeSoundUseCases{},
		LightingUseCases:      world,
		VisualEffectsUseCases: &testutil.FakeVisualEffectsUseCases{},
		ProgressionUseCases:   world,
		EnemyInputAdapter:     world,
		SoundPlayerAdapter:    &testutil.FakeSoundPlayer{},
		MenuInput:             &testutil.FakeMenuInput{},
		RewardAdapter:         &testutil.FakeReward{},
		StageSession:          session_entities.NewStageSessionEntity(),
		BonusesRepository:     world,
		SettingsOverlay:       &SettingsOverlay{},
	})
	state.isSetUp = true

	state.Update()

	if !stageUseCases.paused || state.result == nil {
		t.Fatalf("в кадре завершения: пауза %v, итоги %v",
			stageUseCases.paused, state.result != nil)
	}
}

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
