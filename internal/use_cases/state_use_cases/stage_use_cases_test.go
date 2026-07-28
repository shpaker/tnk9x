package state_use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	state_use_cases "github.com/shpaker/tnk9x/internal/use_cases/state_use_cases"
)

const testDT = 1.0 / 60.0

type spawnEnemyCall struct {
	index     *int
	ignore    bool
	remaining uint
}

// stubLifecycle записывает вызовы спавна; фабрики nextEnemy/nextPlayer*
// позволяют вернуть nil (заблокированный спавнер) или новый танк
type stubLifecycle struct {
	initialEnemies [3]*types.TankEntity
	nextEnemy      func() *types.TankEntity
	nextPlayer1    func() *types.TankEntity
	nextPlayer2    func() *types.TankEntity
	spawnCalls     []spawnEnemyCall
	player1Calls   int
	player2Calls   int
	players        [2]*types.TankEntity
}

func (s *stubLifecycle) OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error) {
	return s.initialEnemies, nil
}

func (s *stubLifecycle) SpawnEnemyWithLevel(
	index *int,
	ignoreRespawnDelay bool,
	remainingEnemies uint,
) (*types.TankEntity, error) {
	s.spawnCalls = append(s.spawnCalls, spawnEnemyCall{
		index:     index,
		ignore:    ignoreRespawnDelay,
		remaining: remainingEnemies,
	})
	if s.nextEnemy == nil {
		return nil, nil
	}
	return s.nextEnemy(), nil
}

func (s *stubLifecycle) SpawnPlayer1() (*types.TankEntity, error) {
	s.player1Calls++
	if s.nextPlayer1 == nil {
		return nil, nil
	}
	tank := s.nextPlayer1()
	s.players[types.PlayerTankNumPlayer1] = tank
	return tank, nil
}

func (s *stubLifecycle) SpawnPlayer2() (*types.TankEntity, error) {
	s.player2Calls++
	if s.nextPlayer2 == nil {
		return nil, nil
	}
	tank := s.nextPlayer2()
	s.players[types.PlayerTankNumPlayer2] = tank
	return tank, nil
}

func (s *stubLifecycle) GetPlayerTank(
	num types.PlayerTankNum,
) *types.TankEntity {
	return s.players[num]
}

func (s *stubLifecycle) SetPlayerTank(
	num types.PlayerTankNum,
	tank *types.TankEntity,
) {
	s.players[num] = tank
}

func (s *stubLifecycle) Explode(tank *types.TankEntity) error { return nil }

func (s *stubLifecycle) UpdateAllTanksLifecycle() error { return nil }

type stubTankCommon struct {
	tanks []*types.TankEntity
}

func (s *stubTankCommon) Update(tank *types.TankEntity, dt float64) error {
	return nil
}
func (s *stubTankCommon) UpdateAllTanks(dt float64) error { return nil }

func (s *stubTankCommon) GetAllTanks() []*types.TankEntity { return s.tanks }

func (s *stubTankCommon) GetAllPlayerTanks() []*types.TankEntity { return nil }

func (s *stubTankCommon) IsAnyPlayerTankMoving() bool      { return false }
func (s *stubTankCommon) LevelUp(tank *types.TankEntity)   {}
func (s *stubTankCommon) LevelDown(tank *types.TankEntity) {}

func (s *stubTankCommon) GetTankAnimationName(
	tank *types.TankEntity,
) string {
	return ""
}

type stubBulletUseCases struct {
	updateCalls int
}

func (s *stubBulletUseCases) ShootBullet(tank *types.TankEntity) error {
	return nil
}

func (s *stubBulletUseCases) UpdateBullets(dt float64) error {
	s.updateCalls++
	return nil
}

func (s *stubBulletUseCases) GetBullets() []*types.BulletEntity { return nil }

func (s *stubBulletUseCases) RemoveBullet(bullet *types.BulletEntity) error {
	return nil
}

type stubCollisionUseCases struct {
	updateCalls int
}

func (s *stubCollisionUseCases) UpdateCollisions() { s.updateCalls++ }

func (s *stubCollisionUseCases) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
) bool {
	return false
}

type stubHQUseCases struct {
	destroyed       bool
	explosionChecks int
}

func (s *stubHQUseCases) GetHQ() *types.HQEntity           { return nil }
func (s *stubHQUseCases) Explode(hq *types.HQEntity) error { return nil }

func (s *stubHQUseCases) IsExplosionFinished(hq *types.HQEntity) {
	s.explosionChecks++
}

func (s *stubHQUseCases) IsDestroyed() bool { return s.destroyed }

type stageTestEnv struct {
	lifecycle *stubLifecycle
	common    *stubTankCommon
	bullets   *stubBulletUseCases
	collision *stubCollisionUseCases
	hq        *stubHQUseCases
	session   *session_entities.StageSessionEntity
	bonuses   *game.BonusesRepository
	stage     *state_use_cases.StageUseCases
}

func newStageTestEnv(enemyRespawnDelay uint) *stageTestEnv {
	lifecycle := &stubLifecycle{}
	common := &stubTankCommon{}
	bullets := &stubBulletUseCases{}
	collision := &stubCollisionUseCases{}
	hq := &stubHQUseCases{}
	session := session_entities.NewStageSessionEntity()
	bonuses := game.NewBonusesRepository()

	// NewStageUseCases возвращает значение — берём адрес
	stage := state_use_cases.NewStageUseCases(
		lifecycle,
		common,
		bullets,
		collision,
		hq,
		session,
		enemyRespawnDelay,
		bonuses,
		nil, // mapUseCases: пути с ним защищены nil-проверками
		nil, // bonusUseCases: пути с ним защищены nil-проверками
	)

	return &stageTestEnv{
		lifecycle: lifecycle,
		common:    common,
		bullets:   bullets,
		collision: collision,
		hq:        hq,
		session:   session,
		bonuses:   bonuses,
		stage:     &stage,
	}
}

func newTankInState(
	role types.TankRole,
	state types.TankState,
) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(role, types.DirectionUp)
	tank := &tankValue
	tank.State = state
	return tank
}

func newEnemyFactory() func() *types.TankEntity {
	return func() *types.TankEntity {
		return newTankInState(types.TankRoleEnemy, types.TankStateSpawning)
	}
}

func (env *stageTestEnv) destroyAllEnemies() {
	total := int(env.session.GetTotalEnemies())
	for i := 0; i < total; i++ {
		env.session.IncrementDestroyedEnemies()
	}
}

func TestStageUseCases_PauseControls(t *testing.T) {
	env := newStageTestEnv(1)

	if env.stage.IsPaused() {
		t.Error("новый уровень не должен быть на паузе")
	}

	env.stage.TogglePause()
	if !env.stage.IsPaused() {
		t.Error("TogglePause не включил паузу")
	}
	env.stage.TogglePause()
	if env.stage.IsPaused() {
		t.Error("TogglePause не выключил паузу")
	}

	env.stage.PauseStageState()
	if !env.stage.IsPaused() {
		t.Error("PauseStageState не включил паузу")
	}
	env.stage.ResumeStageState()
	if env.stage.IsPaused() {
		t.Error("ResumeStageState не выключил паузу")
	}
}

// На паузе игровые объекты не обновляются
func TestStageUseCases_UpdateGameObjects_PausedSkipsWork(t *testing.T) {
	env := newStageTestEnv(1)

	env.stage.PauseStageState()
	env.stage.UpdateGameObjects(testDT)
	env.stage.UpdateGameObjects(testDT)

	if env.bullets.updateCalls != 0 || env.collision.updateCalls != 0 ||
		env.hq.explosionChecks != 0 {
		t.Errorf(
			"на паузе были вызовы: bullets=%d collision=%d hq=%d",
			env.bullets.updateCalls,
			env.collision.updateCalls,
			env.hq.explosionChecks,
		)
	}

	env.stage.ResumeStageState()
	env.stage.UpdateGameObjects(testDT)

	if env.bullets.updateCalls != 1 || env.collision.updateCalls != 1 ||
		env.hq.explosionChecks != 1 {
		t.Errorf(
			"после паузы: bullets=%d collision=%d hq=%d, ожидалось по 1",
			env.bullets.updateCalls,
			env.collision.updateCalls,
			env.hq.explosionChecks,
		)
	}
}

// Таблица истинности победы и поражения
func TestStageUseCases_WinLoseTruthTable(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(env *stageTestEnv)
		wantWon  bool
		wantLost bool
	}{
		{
			name:     "начало уровня",
			setup:    func(env *stageTestEnv) {},
			wantWon:  false,
			wantLost: false,
		},
		{
			name: "все враги уничтожены",
			setup: func(env *stageTestEnv) {
				env.destroyAllEnemies()
			},
			wantWon:  true,
			wantLost: false,
		},
		{
			name: "враги уничтожены, но игроки без жизней",
			setup: func(env *stageTestEnv) {
				env.destroyAllEnemies()
				env.session.SetPlayerLives(types.PlayerTankNumPlayer1, 0)
				env.session.SetPlayerLives(types.PlayerTankNumPlayer2, 0)
			},
			wantWon:  false,
			wantLost: true,
		},
		{
			name: "жив только второй игрок",
			setup: func(env *stageTestEnv) {
				env.destroyAllEnemies()
				env.session.SetPlayerLives(types.PlayerTankNumPlayer1, 0)
			},
			wantWon:  true,
			wantLost: false,
		},
		{
			name: "все игроки без жизней",
			setup: func(env *stageTestEnv) {
				env.session.SetPlayerLives(types.PlayerTankNumPlayer1, 0)
				env.session.SetPlayerLives(types.PlayerTankNumPlayer2, 0)
			},
			wantWon:  false,
			wantLost: true,
		},
		{
			name: "штаб уничтожен",
			setup: func(env *stageTestEnv) {
				env.hq.destroyed = true
			},
			wantWon:  false,
			wantLost: true,
		},
		{
			name: "враги уничтожены, но штаб потерян",
			setup: func(env *stageTestEnv) {
				env.destroyAllEnemies()
				env.hq.destroyed = true
			},
			wantWon:  false,
			wantLost: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newStageTestEnv(1)
			tt.setup(env)

			if got := env.stage.IsStageWon(); got != tt.wantWon {
				t.Errorf("IsStageWon = %v, ожидалось %v", got, tt.wantWon)
			}
			if got := env.stage.IsStageLost(); got != tt.wantLost {
				t.Errorf("IsStageLost = %v, ожидалось %v", got, tt.wantLost)
			}
			wantFinished := tt.wantWon || tt.wantLost
			if got := env.stage.IsStageFinished(); got != wantFinished {
				t.Errorf(
					"IsStageFinished = %v, ожидалось %v",
					got,
					wantFinished,
				)
			}
		})
	}
}

// Без hqUseCases победа определяется только по врагам и игрокам
func TestStageUseCases_IsStageWon_WithoutHQUseCases(t *testing.T) {
	session := session_entities.NewStageSessionEntity()
	stage := state_use_cases.NewStageUseCases(
		nil, nil, nil, nil, nil, session, 0, nil, nil, nil,
	)
	for i := 0; i < int(session.GetTotalEnemies()); i++ {
		session.IncrementDestroyedEnemies()
	}

	if !stage.IsStageWon() {
		t.Error("ожидалась победа без hqUseCases")
	}
	if stage.IsStageLost() {
		t.Error("поражение при живых игроках без hqUseCases")
	}
}

// Все зависимости nil: методы не паникуют и возвращают нейтральные значения
func TestStageUseCases_NilDependenciesAreSafe(t *testing.T) {
	stage := state_use_cases.NewStageUseCases(
		nil, nil, nil, nil, nil, nil, 0, nil, nil, nil,
	)

	if stage.IsStageWon() || stage.IsStageLost() || stage.IsStageFinished() {
		t.Error("без сессии уровень не выигран и не проигран")
	}
	if got := stage.TrySpawnEnemy(); got != nil {
		t.Errorf("TrySpawnEnemy: %v", got)
	}
	if r1, r2 := stage.TryRespawnPlayersTanks(); r1 != nil || r2 != nil {
		t.Errorf("TryRespawnPlayersTanks: %v, %v", r1, r2)
	}
	if got := stage.SpawnPlayerTank(types.TankRolePlayer1); got != nil {
		t.Errorf("SpawnPlayerTank: %v", got)
	}
	if got := stage.SpawnInitialEnemyTanks(); got != nil {
		t.Errorf("SpawnInitialEnemyTanks: %v", got)
	}
	if got := stage.GetPlayersTanks(); len(got) != 2 ||
		got[0] != nil || got[1] != nil {
		t.Errorf("GetPlayersTanks: %v", got)
	}
	stage.UpdateGameObjects(testDT) // не должно паниковать
}

// Спавн врага только после отсчёта enemyRespawnDelay тиков
func TestStageUseCases_TrySpawnEnemy_RespawnDelay(t *testing.T) {
	env := newStageTestEnv(2)
	env.lifecycle.nextEnemy = newEnemyFactory()

	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("враг заспавнился до отсчёта")
	}
	env.stage.UpdateGameObjects(testDT)
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("враг заспавнился на середине отсчёта")
	}
	env.stage.UpdateGameObjects(testDT)

	spawned := env.stage.TrySpawnEnemy()
	if spawned == nil {
		t.Fatal("враг не заспавнился после отсчёта")
	}
	if len(env.lifecycle.spawnCalls) != 1 {
		t.Fatalf("вызовов спавна %d", len(env.lifecycle.spawnCalls))
	}
	call := env.lifecycle.spawnCalls[0]
	if call.index != nil || call.ignore || call.remaining != 20 {
		t.Errorf("аргументы спавна: %+v", call)
	}

	// После спавна отсчёт начинается заново
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Error("враг заспавнился без нового отсчёта")
	}
	if got := env.session.GetNextEnemyNumber(); got != 2 {
		t.Errorf("следующий номер врага %d, ожидался 2", got)
	}
}

// Нулевая задержка в конструкторе заменяется на 3*60 тиков
func TestStageUseCases_TrySpawnEnemy_DefaultDelay(t *testing.T) {
	env := newStageTestEnv(0)
	env.lifecycle.nextEnemy = newEnemyFactory()

	for i := 0; i < 180; i++ {
		if got := env.stage.TrySpawnEnemy(); got != nil {
			t.Fatalf("враг заспавнился раньше времени на тике %d", i)
		}
		env.stage.UpdateGameObjects(testDT)
	}

	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("враг не заспавнился после 180 тиков")
	}
}

// Лимит одновременно активных врагов (по умолчанию 5)
func TestStageUseCases_TrySpawnEnemy_MaxActiveEnemiesCap(t *testing.T) {
	env := newStageTestEnv(1)
	env.lifecycle.nextEnemy = newEnemyFactory()

	for i := 0; i < 5; i++ {
		env.common.tanks = append(
			env.common.tanks,
			newTankInState(types.TankRoleEnemy, types.TankStateStopped),
		)
	}

	env.stage.UpdateGameObjects(testDT)
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("враг заспавнился при заполненном лимите")
	}
	if len(env.lifecycle.spawnCalls) != 0 {
		t.Fatal("lifecycle вызван при заполненном лимите")
	}

	// Взорванные враги и игроки не считаются активными врагами
	env.common.tanks[0].State = types.TankStateExploded
	env.common.tanks = append(
		env.common.tanks,
		newTankInState(types.TankRolePlayer1, types.TankStateMoving),
	)
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("враг не заспавнился при свободном слоте")
	}
}

// Неудавшийся спавн не сдвигает номер врага и не сбрасывает отсчёт
func TestStageUseCases_TrySpawnEnemy_FailedSpawnKeepsSchedule(t *testing.T) {
	env := newStageTestEnv(1)

	env.stage.UpdateGameObjects(testDT)
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("nextEnemy nil: спавн должен вернуть nil")
	}
	if len(env.lifecycle.spawnCalls) != 1 {
		t.Fatalf("вызовов спавна %d, ожидался 1", len(env.lifecycle.spawnCalls))
	}
	if got := env.session.GetNextEnemyNumber(); got != 1 {
		t.Errorf("номер врага сдвинулся: %d", got)
	}

	// Отсчёт не сброшен: повторная попытка проходит без ожидания
	env.lifecycle.nextEnemy = newEnemyFactory()
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("повторный спавн не удался")
	}
	if got := env.session.GetNextEnemyNumber(); got != 2 {
		t.Errorf("номер врага после спавна %d, ожидался 2", got)
	}
}

// Бонусные враги: номера 4, 9, 15 (первый 4, далее +5, +6, +7...)
func TestStageUseCases_BonusEnemySequence(t *testing.T) {
	env := newStageTestEnv(1)
	env.lifecycle.nextEnemy = newEnemyFactory()

	bonusNumbers := map[uint]bool{4: true, 9: true, 15: true}

	for number := uint(1); number <= 20; number++ {
		env.stage.UpdateGameObjects(testDT)
		tank := env.stage.TrySpawnEnemy()
		if tank == nil {
			t.Fatalf("враг %d не заспавнился", number)
		}
		if tank.GetWithBonus() != bonusNumbers[number] {
			t.Errorf(
				"враг %d: withBonus=%v, ожидалось %v",
				number,
				tank.GetWithBonus(),
				bonusNumbers[number],
			)
		}
	}

	// Все 20 врагов заспавнены — дальше спавн невозможен
	env.stage.UpdateGameObjects(testDT)
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Error("спавн после исчерпания врагов")
	}
}

// Начальный спавн: nil-слоты пропускаются, номера 1-3 без бонусов
func TestStageUseCases_SpawnInitialEnemyTanks(t *testing.T) {
	env := newStageTestEnv(1)
	enemy1 := newTankInState(types.TankRoleEnemy, types.TankStateSpawning)
	enemy2 := newTankInState(types.TankRoleEnemy, types.TankStateSpawning)
	env.lifecycle.initialEnemies = [3]*types.TankEntity{enemy1, enemy2, nil}

	spawned := env.stage.SpawnInitialEnemyTanks()

	if len(spawned) != 2 {
		t.Fatalf("заспавнено %d, ожидалось 2", len(spawned))
	}
	for i, tank := range spawned {
		if tank.GetWithBonus() {
			t.Errorf("начальный враг %d получил бонус", i)
		}
	}
	if got := env.session.GetNextEnemyNumber(); got != 3 {
		t.Errorf("следующий номер врага %d, ожидался 3", got)
	}
}

func TestStageUseCases_SpawnPlayerTank(t *testing.T) {
	env := newStageTestEnv(1)
	playerTank := newTankInState(
		types.TankRolePlayer1,
		types.TankStateSpawning,
	)
	env.lifecycle.nextPlayer1 = func() *types.TankEntity { return playerTank }

	if got := env.stage.SpawnPlayerTank(types.TankRolePlayer1); got != playerTank {
		t.Errorf("спавн игрока 1: %v", got)
	}
	if env.lifecycle.player1Calls != 1 {
		t.Errorf("вызовов SpawnPlayer1: %d", env.lifecycle.player1Calls)
	}

	// Роль врага не обслуживается
	if got := env.stage.SpawnPlayerTank(types.TankRoleEnemy); got != nil {
		t.Errorf("спавн по роли врага: %v", got)
	}

	// Побеждённый игрок не спавнится, lifecycle не вызывается
	env.session.SetPlayerLives(types.PlayerTankNumPlayer1, 0)
	if got := env.stage.SpawnPlayerTank(types.TankRolePlayer1); got != nil {
		t.Errorf("спавн побеждённого игрока: %v", got)
	}
	if env.lifecycle.player1Calls != 1 {
		t.Errorf(
			"lifecycle вызван для побеждённого игрока: %d",
			env.lifecycle.player1Calls,
		)
	}
}

// Респавн взорванного игрока списывает жизнь
func TestStageUseCases_TryRespawnPlayersTanks_Respawn(t *testing.T) {
	env := newStageTestEnv(1)
	exploded := newTankInState(
		types.TankRolePlayer1,
		types.TankStateExploded,
	)
	env.lifecycle.players[types.PlayerTankNumPlayer1] = exploded
	fresh := newTankInState(types.TankRolePlayer1, types.TankStateSpawning)
	env.lifecycle.nextPlayer1 = func() *types.TankEntity { return fresh }

	respawned1, respawned2 := env.stage.TryRespawnPlayersTanks()

	if respawned1 != fresh {
		t.Errorf("respawned1 = %v, ожидался новый танк", respawned1)
	}
	if respawned2 != nil {
		t.Errorf("respawned2 = %v", respawned2)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 2 {
		t.Errorf("жизни игрока 1: %d, ожидалось 2", got)
	}
}

// Заблокированный спавн возвращает списанную жизнь
func TestStageUseCases_TryRespawnPlayersTanks_BlockedRestoresLife(
	t *testing.T,
) {
	env := newStageTestEnv(1)
	exploded := newTankInState(
		types.TankRolePlayer1,
		types.TankStateExploded,
	)
	env.lifecycle.players[types.PlayerTankNumPlayer1] = exploded
	// nextPlayer1 nil: lifecycle вернёт nil-танк (спавнер заблокирован)

	respawned1, _ := env.stage.TryRespawnPlayersTanks()

	if respawned1 != nil {
		t.Errorf("respawned1 = %v", respawned1)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 3 {
		t.Errorf("жизнь не возвращена: %d, ожидалось 3", got)
	}
}

// Последняя жизнь: игрок становится побеждённым, жизнь не возвращается
func TestStageUseCases_TryRespawnPlayersTanks_LastLife(t *testing.T) {
	env := newStageTestEnv(1)
	env.session.SetPlayerLives(types.PlayerTankNumPlayer1, 1)
	exploded := newTankInState(
		types.TankRolePlayer1,
		types.TankStateExploded,
	)
	env.lifecycle.players[types.PlayerTankNumPlayer1] = exploded
	fresh := newTankInState(types.TankRolePlayer1, types.TankStateSpawning)
	env.lifecycle.nextPlayer1 = func() *types.TankEntity { return fresh }

	respawned1, _ := env.stage.TryRespawnPlayersTanks()

	if respawned1 != nil {
		t.Errorf("респавн с последней жизни: %v", respawned1)
	}
	if env.lifecycle.player1Calls != 0 {
		t.Errorf("lifecycle вызван: %d", env.lifecycle.player1Calls)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 0 {
		t.Errorf("жизни: %d, ожидалось 0", got)
	}
	if !env.session.IsPlayerDefeated(types.PlayerTankNumPlayer1) {
		t.Error("игрок не помечен побеждённым")
	}
}

// Респавнятся только взорванные танки, при одном игроке второй не трогается
func TestStageUseCases_TryRespawnPlayersTanks_OnlyExploded(t *testing.T) {
	env := newStageTestEnv(1)
	alive := newTankInState(types.TankRolePlayer1, types.TankStateStopped)
	env.lifecycle.players[types.PlayerTankNumPlayer1] = alive
	// Второй игрок взорван, но playerCount=1 — он вне игры
	env.lifecycle.players[types.PlayerTankNumPlayer2] = newTankInState(
		types.TankRolePlayer2,
		types.TankStateExploded,
	)
	env.lifecycle.nextPlayer2 = func() *types.TankEntity {
		return newTankInState(types.TankRolePlayer2, types.TankStateSpawning)
	}

	respawned1, respawned2 := env.stage.TryRespawnPlayersTanks()

	if respawned1 != nil || respawned2 != nil {
		t.Errorf("неожиданный респавн: %v, %v", respawned1, respawned2)
	}
	if env.lifecycle.player2Calls != 0 {
		t.Errorf(
			"SpawnPlayer2 вызван при playerCount=1: %d",
			env.lifecycle.player2Calls,
		)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 3 {
		t.Errorf("жизни живого игрока изменились: %d", got)
	}
}

func TestStageUseCases_TryRespawnPlayersTanks_TwoPlayers(t *testing.T) {
	env := newStageTestEnv(1)
	env.session.SetPlayerCount(2)
	env.lifecycle.players[types.PlayerTankNumPlayer1] = newTankInState(
		types.TankRolePlayer1,
		types.TankStateExploded,
	)
	env.lifecycle.players[types.PlayerTankNumPlayer2] = newTankInState(
		types.TankRolePlayer2,
		types.TankStateExploded,
	)
	fresh1 := newTankInState(types.TankRolePlayer1, types.TankStateSpawning)
	fresh2 := newTankInState(types.TankRolePlayer2, types.TankStateSpawning)
	env.lifecycle.nextPlayer1 = func() *types.TankEntity { return fresh1 }
	env.lifecycle.nextPlayer2 = func() *types.TankEntity { return fresh2 }

	respawned1, respawned2 := env.stage.TryRespawnPlayersTanks()

	if respawned1 != fresh1 || respawned2 != fresh2 {
		t.Errorf("респавн: %v, %v", respawned1, respawned2)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 2 {
		t.Errorf("жизни игрока 1: %d", got)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer2); got != 2 {
		t.Errorf("жизни игрока 2: %d", got)
	}
}

// Каждый уничтоженный враг учитывается ровно один раз
func TestStageUseCases_TrackDestroyedEnemies(t *testing.T) {
	env := newStageTestEnv(1)
	deadEnemy := newTankInState(
		types.TankRoleEnemy,
		types.TankStateExploded,
	)
	aliveEnemy := newTankInState(
		types.TankRoleEnemy,
		types.TankStateStopped,
	)
	deadPlayer := newTankInState(
		types.TankRolePlayer1,
		types.TankStateExploded,
	)
	env.common.tanks = []*types.TankEntity{deadEnemy, aliveEnemy, deadPlayer}

	env.stage.UpdateGameObjects(testDT)
	if got := env.session.GetRemainingEnemies(); got != 19 {
		t.Fatalf("осталось врагов %d, ожидалось 19", got)
	}

	// Повторные тики не считают того же врага снова
	env.stage.UpdateGameObjects(testDT)
	env.stage.UpdateGameObjects(testDT)
	if got := env.session.GetRemainingEnemies(); got != 19 {
		t.Errorf("враг посчитан повторно: осталось %d", got)
	}

	aliveEnemy.State = types.TankStateExploded
	env.stage.UpdateGameObjects(testDT)
	if got := env.session.GetRemainingEnemies(); got != 18 {
		t.Errorf("второй враг не посчитан: осталось %d", got)
	}
}

// SpawnInitialEnemyTanks очищает учёт уничтоженных врагов:
// тот же взорванный танк учитывается заново
func TestStageUseCases_TrackDestroyedEnemies_ResetOnInitialSpawn(
	t *testing.T,
) {
	env := newStageTestEnv(1)
	deadEnemy := newTankInState(
		types.TankRoleEnemy,
		types.TankStateExploded,
	)
	env.common.tanks = []*types.TankEntity{deadEnemy}

	env.stage.UpdateGameObjects(testDT)
	if got := env.session.GetRemainingEnemies(); got != 19 {
		t.Fatalf("осталось врагов %d, ожидалось 19", got)
	}

	env.stage.SpawnInitialEnemyTanks()
	env.stage.UpdateGameObjects(testDT)
	if got := env.session.GetRemainingEnemies(); got != 18 {
		t.Errorf(
			"после сброса учёта осталось %d, ожидалось 18",
			got,
		)
	}
}

// Уничтожение бонусного врага удаляет бонусы без владельца
func TestStageUseCases_TrackDestroyedEnemies_BonusEnemyClearsBonuses(
	t *testing.T,
) {
	env := newStageTestEnv(1)
	bonusEnemy := newTankInState(
		types.TankRoleEnemy,
		types.TankStateExploded,
	)
	bonusEnemy.SetWithBonus(true)
	env.common.tanks = []*types.TankEntity{bonusEnemy}

	ownerless := types.NewBonusEntity(
		types.BonusTypeStar,
		types.Position{},
		types.Size{Width: 16, Height: 16},
		nil,
	)
	owned := types.NewBonusEntity(
		types.BonusTypeTank,
		types.Position{},
		types.Size{Width: 16, Height: 16},
		nil,
	)
	owned.SetOwner(newTankInState(
		types.TankRolePlayer1,
		types.TankStateStopped,
	))
	env.bonuses.AddBonus(ownerless)
	env.bonuses.AddBonus(owned)

	env.stage.UpdateGameObjects(testDT)

	remaining := env.bonuses.GetAllBonuses()
	if len(remaining) != 1 || remaining[0] != owned {
		t.Errorf("бонусы после уничтожения: %v", remaining)
	}
}

func TestStageUseCases_GetPlayersTanks(t *testing.T) {
	env := newStageTestEnv(1)
	tank1 := newTankInState(types.TankRolePlayer1, types.TankStateStopped)
	env.lifecycle.players[types.PlayerTankNumPlayer1] = tank1

	tanks := env.stage.GetPlayersTanks()
	if len(tanks) != 2 {
		t.Fatalf("длина %d, ожидалось 2", len(tanks))
	}
	if tanks[0] != tank1 || tanks[1] != nil {
		t.Errorf("танки игроков: %v", tanks)
	}
}
