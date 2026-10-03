package states

import (
	"slices"
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

func (f *fakeFinishingStageUseCases) UpdateGameObjects(
	float64,
) {
	f.finished = true
}

func (f *fakeFinishingStageUseCases) IsStageFinished() bool { return f.finished }
func (f *fakeFinishingStageUseCases) IsStageWon() bool      { return false }

func (f *fakeFinishingStageUseCases) PauseStageState()       { f.paused = true }
func (f *fakeFinishingStageUseCases) CanRevivePlayers() bool { return false }

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

func (fakeStageWorld) UpdateAllTanksLifecycle() error { return nil }
func (fakeStageWorld) UpdateAnimations()              {}
func (fakeStageWorld) UpdateHeadlights()              {}
func (fakeStageWorld) Update(float64)                 {}

func (fakeStageWorld) GetAllBonuses() []*types.BonusEntity { return nil }

func (fakeStageWorld) CalcStars(
	types.StageResult,
	*types.LevelEntity,
) uint {
	return 0
}

func (fakeStageWorld) GetLevelStars(
	int,
) uint {
	return 0
}

func (fakeStageWorld) NextLevel(
	level int,
) (int, bool) {
	return level + 1, false
}

// Итоги строятся в том же кадре, где уровень завершился: иначе Draw
// видит паузу без итогов и на кадр рисует меню паузы
func TestStageState_FinishBuildsResultSameFrame(t *testing.T) {
	stageUseCases := &fakeFinishingStageUseCases{}
	world := fakeStageWorld{}
	state := NewStageState(StageStateDependencies{
		InventoryUseCases:     &testutil.FakeInventory{},
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

// fakeStageRenderer — пункт меню под точкой равен её Y
type fakeStageRenderer struct {
	StageRenderer
}

func (fakeStageRenderer) HitPauseRow(position types.Position) (int, bool) {
	return int(position.Y), true
}

func (fakeStageRenderer) HitResultRow(position types.Position) (int, bool) {
	return int(position.Y), true
}

func TestStageState_PauseStopsEngine(t *testing.T) {
	stageUseCases := &fakeStageUseCases{}
	soundUseCases := &fakeSoundUseCases{}
	input := &testutil.FakeMenuInput{}
	state := NewStageState(StageStateDependencies{
		InventoryUseCases: &testutil.FakeInventory{},
		StageUseCases:     stageUseCases,
		SoundUseCases:     soundUseCases,
		Renderer:          fakeStageRenderer{},
		MenuInput:         input,
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
		t.Errorf(
			"остановка двигателя — один раз на открытие, их %d",
			engineStops(),
		)
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
		t.Errorf(
			"каждое открытие меню глушит двигатель, остановок %d",
			engineStops(),
		)
	}
}

func TestStageState_PauseMenuMouse(t *testing.T) {
	stageUseCases := &fakeStageUseCases{}
	input := &testutil.FakeMenuInput{}
	state := NewStageState(StageStateDependencies{
		InventoryUseCases: &testutil.FakeInventory{},
		StageUseCases:     stageUseCases,
		SoundUseCases:     &fakeSoundUseCases{},
		Renderer:          fakeStageRenderer{},
		MenuInput:         input,
	})
	frame := func(pressed testutil.FakeMenuInput) {
		*input = pressed
		state.handlePauseMenu()
		input.Reset()
	}

	// Правая кнопка в бою паузу не ставит
	frame(testutil.FakeMenuInput{BackPressed: true})
	if stageUseCases.paused {
		t.Fatal("назад в бою не ставит паузу")
	}

	// Наведение выделяет пункт, клик по CONTINUE снимает паузу
	frame(testutil.FakeMenuInput{Pause: true})
	frame(testutil.FakeMenuInput{
		PointPosition: types.Position{Y: 2}, PointMoved: true,
	})
	if state.pauseMenuIndex != 2 {
		t.Errorf("наведение выделяет пункт, курсор %d", state.pauseMenuIndex)
	}
	frame(testutil.FakeMenuInput{
		TapPosition: types.Position{Y: 0}, TapPressed: true,
	})
	if stageUseCases.paused {
		t.Error("клик по CONTINUE снимает паузу")
	}
}

// fakeBoostingStageUseCases — поражение с усиленным переносом
type fakeBoostingStageUseCases struct {
	fakeFinishingStageUseCases
	boosts int
}

func (f *fakeBoostingStageUseCases) BoostCarryOver() { f.boosts++ }

// С жетонами пункты за рекламу есть и без рекламы площадки,
// а выбор пункта тратит жетон вместо показа рекламы
func TestStageState_TokensPayForReward(t *testing.T) {
	stageUseCases := &fakeBoostingStageUseCases{}
	inventory := &testutil.FakeInventory{Tokens: 1}
	reward := &testutil.FakeReward{}
	world := fakeStageWorld{}
	state := NewStageState(StageStateDependencies{
		InventoryUseCases:     inventory,
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
		RewardAdapter:         reward,
		StageSession:          session_entities.NewStageSessionEntity(),
		BonusesRepository:     world,
		SettingsOverlay:       &SettingsOverlay{},
	})
	state.isSetUp = true
	state.Update()

	if state.result.Tokens != 1 ||
		!slices.Contains(state.result.Items, types.StageResultItemBoostRetry) {
		t.Fatalf("result %+v must offer BOOST for a token", state.result)
	}

	transition := state.applyStageResultItem(types.StageResultItemBoostRetry)
	if transition.Target != types.TransitionToStage || !transition.CarryOver {
		t.Errorf("transition %+v, want boosted retry", transition)
	}
	if inventory.Tokens != 0 || reward.Requests != 0 ||
		stageUseCases.boosts != 1 {
		t.Errorf("tokens %d, ad requests %d, boosts %d",
			inventory.Tokens, reward.Requests, stageUseCases.boosts)
	}
}

// newResultStageState — уровень, завершающийся поражением в первом
// же кадре, с заданными жетонами и продажей жетонов
func newResultStageState(
	inventory *testutil.FakeInventory,
	reward *testutil.FakeReward,
	tokensOnSale bool,
	input *testutil.FakeMenuInput,
) *StageState {
	world := fakeStageWorld{}
	state := NewStageState(StageStateDependencies{
		InventoryUseCases:     inventory,
		TankCommonUseCases:    &testutil.FakeTankCommonUseCases{},
		TankLifecycleUseCases: world,
		TilesUseCases:         world,
		StageUseCases:         &fakeFinishingStageUseCases{},
		SoundUseCases:         &fakeSoundUseCases{},
		LightingUseCases:      world,
		VisualEffectsUseCases: &testutil.FakeVisualEffectsUseCases{},
		ProgressionUseCases:   world,
		EnemyInputAdapter:     world,
		Renderer:              fakeStageRenderer{},
		SoundPlayerAdapter:    &testutil.FakeSoundPlayer{},
		MenuInput:             input,
		RewardAdapter:         reward,
		TokensOnSale:          tokensOnSale,
		StageSession:          session_entities.NewStageSessionEntity(),
		BonusesRepository:     world,
		SettingsOverlay:       &SettingsOverlay{},
	})
	state.isSetUp = true
	return state
}

// Подсказка про жетоны — только когда их нет, площадка их продаёт
// и в меню есть пункты за рекламу
func TestStageState_TokensHint(t *testing.T) {
	tests := []struct {
		name         string
		tokens       uint
		adAvailable  bool
		tokensOnSale bool
		want         bool
	}{
		{"нет жетонов, продаются", 0, true, true, true},
		{"жетоны есть", 2, true, true, false},
		{"жетоны не продаются", 0, true, false, false},
		{"нет пунктов за рекламу", 0, false, true, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := newResultStageState(
				&testutil.FakeInventory{Tokens: test.tokens},
				&testutil.FakeReward{Available: test.adAvailable},
				test.tokensOnSale,
				&testutil.FakeMenuInput{},
			)
			state.Update()

			if state.result.TokensHint != test.want {
				t.Errorf("hint %v, want %v", state.result.TokensHint, test.want)
			}
			if state.result.Tokens != test.tokens {
				t.Errorf("tokens %d, want %d", state.result.Tokens, test.tokens)
			}
		})
	}
}

// Уход с итогов выигранного уровня — победа, с проигранного — нет
func TestStageState_ResultTransitionVictory(t *testing.T) {
	for _, won := range []bool{true, false} {
		input := &testutil.FakeMenuInput{}
		state := newResultStageState(
			&testutil.FakeInventory{}, &testutil.FakeReward{}, false, input,
		)
		state.result = &types.StageResultViewData{
			Won:   won,
			Items: []types.StageResultItem{types.StageResultItemLevels},
		}
		// Экран итогов уже появился целиком
		state.resultTicks = resultRevealLength(0)

		if transition := state.handleStageResult(); transition.Victory {
			t.Errorf("won %v: no choice yet, no victory", won)
		}
		input.BackPressed = true
		transition := state.handleStageResult()
		if transition.Target != types.TransitionToLevelSelect ||
			transition.Victory != won {
			t.Errorf("won %v: transition %+v", won, transition)
		}
	}
}
