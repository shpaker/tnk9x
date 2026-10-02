package input_adapters

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

// recordingTankActions записывает вызовы команд танка
type recordingTankActions struct {
	rotations []types.Direction
	moves     int
	stops     int
	shoots    int
}

func (r *recordingTankActions) Update(
	*types.TankEntity, float64,
) error {
	return nil
}

func (r *recordingTankActions) Rotate(
	_ *types.TankEntity,
	direction types.Direction,
) error {
	r.rotations = append(r.rotations, direction)

	return nil
}

func (r *recordingTankActions) Move(*types.TankEntity) error {
	r.moves++

	return nil
}

func (r *recordingTankActions) Stop(*types.TankEntity, bool) {
	r.stops++
}

func (r *recordingTankActions) Shoot(*types.TankEntity) error {
	r.shoots++

	return nil
}

func (r *recordingTankActions) ApplyDecision(
	*types.TankEntity, types.EnemyAIDecision,
) {
}
func (r *recordingTankActions) SetMinXPosition(*types.TankEntity) {}
func (r *recordingTankActions) SetMaxXPosition(*types.TankEntity) {}
func (r *recordingTankActions) SetMinYPosition(*types.TankEntity) {}
func (r *recordingTankActions) SetMaxYPosition(*types.TankEntity) {}

// stubStageUseCases — минимальный стаб паузы уровня
type stubStageUseCases struct {
	paused       bool
	pauseToggles int
}

func (s *stubStageUseCases) SpawnPlayerTank(
	types.TankRole,
) *types.TankEntity {
	return nil
}

func (s *stubStageUseCases) PlacePlayerTank(
	types.TankRole,
) *types.TankEntity {
	return nil
}

func (s *stubStageUseCases) SpawnInitialEnemyTanks() []*types.TankEntity {
	return nil
}
func (s *stubStageUseCases) TrySpawnEnemy() *types.TankEntity { return nil }
func (s *stubStageUseCases) TryRespawnPlayersTanks() (
	*types.TankEntity, *types.TankEntity,
) {
	return nil, nil
}

func (s *stubStageUseCases) GetPlayersTanks() []*types.TankEntity {
	return nil
}
func (s *stubStageUseCases) UpdateGameObjects(float64) {}

func (s *stubStageUseCases) TogglePause() {
	s.pauseToggles++
	s.paused = !s.paused
}

func (s *stubStageUseCases) IsPaused() bool        { return s.paused }
func (s *stubStageUseCases) PauseStageState()      {}
func (s *stubStageUseCases) ResumeStageState()     {}
func (s *stubStageUseCases) IsStageWon() bool      { return false }
func (s *stubStageUseCases) IsStageLost() bool     { return false }
func (s *stubStageUseCases) IsStageFinished() bool { return false }

func (s *stubStageUseCases) GetStageResult() types.StageResult {
	return types.StageResult{}
}

func (s *stubStageUseCases) BoostCarryOver()        {}
func (s *stubStageUseCases) CanRevivePlayers() bool { return false }
func (s *stubStageUseCases) RevivePlayers()         {}

func (s *stubStageUseCases) SaveCarryOver() {}

func newTouchAdapterUnderTest() (
	*StageTouchInputAdapter,
	*testutil.FakeTouchControls,
	*recordingTankActions,
	*stubStageUseCases,
) {
	touch := &testutil.FakeTouchControls{}
	actions := &recordingTankActions{}
	stage := &stubStageUseCases{}
	adapter := NewStageTouchInputAdapter(
		actions, &types.TankEntity{}, stage, touch,
		types.PlayerTankNumPlayer1,
	)

	return adapter, touch, actions, stage
}

func TestStageTouchInputAdapter_SteeringAndStop(t *testing.T) {
	adapter, touch, actions, _ := newTouchAdapterUnderTest()

	// Удержание крестовины: Rotate+Move каждый кадр
	touch.Directions[0] = types.DirectionLeft
	touch.HasDirection[0] = true
	adapter.Update(0)
	adapter.Update(0)
	if actions.moves != 2 || len(actions.rotations) != 2 {
		t.Errorf(
			"ожидалось 2 Rotate+Move, получено %d/%d",
			len(actions.rotations), actions.moves,
		)
	}
	if actions.rotations[0] != types.DirectionLeft {
		t.Error("направление должно передаваться в Rotate")
	}

	// Отпускание: ровно один Stop
	touch.HasDirection[0] = false
	adapter.Update(0)
	adapter.Update(0)
	if actions.stops != 1 {
		t.Errorf("ожидался один Stop, получено %d", actions.stops)
	}
}

func TestStageTouchInputAdapter_ShootAndPause(t *testing.T) {
	adapter, touch, actions, stage := newTouchAdapterUnderTest()

	touch.FireJust[0] = true
	adapter.Update(0)
	if actions.shoots != 1 {
		t.Errorf("ожидался один выстрел, получено %d", actions.shoots)
	}

	// Паузу переключает стейт уровня, а не адаптер; во время паузы
	// стрельба и движение подавляются
	touch.PauseJust = true
	stage.paused = true
	touch.Directions[0] = types.DirectionUp
	touch.HasDirection[0] = true
	adapter.Update(0)
	if stage.pauseToggles != 0 {
		t.Errorf("адаптер не должен переключать паузу: %d", stage.pauseToggles)
	}
	if actions.shoots != 1 || actions.moves != 0 {
		t.Error("во время паузы стрельба и движение подавляются")
	}
}

// Адаптер игрока слушает только свои контроллы
func TestStageTouchInputAdapter_OwnPlayerOnly(t *testing.T) {
	touch := &testutil.FakeTouchControls{}
	actions := &recordingTankActions{}
	adapter := NewStageTouchInputAdapter(
		actions, &types.TankEntity{}, &stubStageUseCases{}, touch,
		types.PlayerTankNumPlayer2,
	)

	touch.FireJust[0] = true
	touch.HasDirection[0] = true
	adapter.Update(0)
	if actions.shoots != 0 || actions.moves != 0 {
		t.Error("контроллы P1 не управляют танком P2")
	}

	touch.FireJust[1] = true
	adapter.Update(0)
	if actions.shoots != 1 {
		t.Error("огонь P2 должен стрелять танком P2")
	}
}

func TestStageTouchInputAdapter_NoTankIsSafe(t *testing.T) {
	touch := &testutil.FakeTouchControls{}
	touch.FireJust[0] = true
	touch.HasDirection[0] = true
	actions := &recordingTankActions{}
	adapter := NewStageTouchInputAdapter(
		actions, nil, &stubStageUseCases{}, touch,
		types.PlayerTankNumPlayer1,
	)

	adapter.Update(0)
	if actions.shoots != 0 || actions.moves != 0 {
		t.Error("без танка команды не должны выполняться")
	}
}
