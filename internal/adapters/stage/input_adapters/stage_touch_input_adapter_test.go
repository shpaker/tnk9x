package input_adapters

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// stubTouchControls — управляемый источник тач-событий
type stubTouchControls struct {
	direction    types.Direction
	hasDirection bool
	fireJust     bool
	pauseJust    bool
}

func (s *stubTouchControls) Update()             {}
func (s *stubTouchControls) IsTouchActive() bool { return true }

func (s *stubTouchControls) DPadDirection() (types.Direction, bool) {
	return s.direction, s.hasDirection
}

func (s *stubTouchControls) FireJustPressed() bool  { return s.fireJust }
func (s *stubTouchControls) PauseJustPressed() bool { return s.pauseJust }

func (s *stubTouchControls) TapJustPressed() (types.Position, bool) {
	return types.Position{}, false
}

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

func newTouchAdapterUnderTest() (
	*StageTouchInputAdapter,
	*stubTouchControls,
	*recordingTankActions,
	*stubStageUseCases,
) {
	touch := &stubTouchControls{}
	actions := &recordingTankActions{}
	stage := &stubStageUseCases{}
	adapter := NewStageTouchInputAdapter(
		actions, &types.TankEntity{}, stage, touch,
	)

	return adapter, touch, actions, stage
}

func TestStageTouchInputAdapter_SteeringAndStop(t *testing.T) {
	adapter, touch, actions, _ := newTouchAdapterUnderTest()

	// Удержание крестовины: Rotate+Move каждый кадр
	touch.direction = types.DirectionLeft
	touch.hasDirection = true
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
	touch.hasDirection = false
	adapter.Update(0)
	adapter.Update(0)
	if actions.stops != 1 {
		t.Errorf("ожидался один Stop, получено %d", actions.stops)
	}
}

func TestStageTouchInputAdapter_ShootAndPause(t *testing.T) {
	adapter, touch, actions, stage := newTouchAdapterUnderTest()

	touch.fireJust = true
	adapter.Update(0)
	if actions.shoots != 1 {
		t.Errorf("ожидался один выстрел, получено %d", actions.shoots)
	}

	// Пауза: тап по паузе переключает её, во время паузы стрельба
	// и движение подавляются
	touch.pauseJust = true
	touch.direction = types.DirectionUp
	touch.hasDirection = true
	adapter.Update(0)
	if stage.pauseToggles != 1 {
		t.Errorf(
			"ожидался один TogglePause, получено %d",
			stage.pauseToggles,
		)
	}
	if actions.shoots != 1 || actions.moves != 0 {
		t.Error("во время паузы стрельба и движение подавляются")
	}
}

func TestStageTouchInputAdapter_NoTankIsSafe(t *testing.T) {
	touch := &stubTouchControls{fireJust: true, hasDirection: true}
	actions := &recordingTankActions{}
	adapter := NewStageTouchInputAdapter(
		actions, nil, &stubStageUseCases{}, touch,
	)

	adapter.Update(0)
	if actions.shoots != 0 || actions.moves != 0 {
		t.Error("без танка команды не должны выполняться")
	}
}
