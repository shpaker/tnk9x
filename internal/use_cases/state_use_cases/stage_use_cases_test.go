package state_use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
	state_use_cases "github.com/shpaker/tnk9x/internal/use_cases/state_use_cases"
)

const testDT = 1.0 / 60.0

type spawnEnemyCall struct {
	spawnerIndex int
	level        uint
}

// stubLifecycle записывает вызовы спавна; фабрики nextEnemy/nextPlayer*
// позволяют вернуть nil (заблокированный спавнер) или новый танк
type stubLifecycle struct {
	nextEnemy    func() *types.TankEntity
	nextPlayer1  func() *types.TankEntity
	nextPlayer2  func() *types.TankEntity
	spawnCalls   []spawnEnemyCall
	player1Calls int
	player1Level uint
	player2Calls int
	players      [2]*types.TankEntity
}

func (s *stubLifecycle) SpawnEnemy(
	spawnerIndex int,
	level uint,
) (*types.TankEntity, error) {
	s.spawnCalls = append(s.spawnCalls, spawnEnemyCall{
		spawnerIndex: spawnerIndex,
		level:        level,
	})
	if s.nextEnemy == nil {
		return nil, nil
	}
	return s.nextEnemy(), nil
}

func (s *stubLifecycle) SpawnPlayer1(level uint) (*types.TankEntity, error) {
	s.player1Calls++
	s.player1Level = level
	if s.nextPlayer1 == nil {
		return nil, nil
	}
	tank := s.nextPlayer1()
	s.players[types.PlayerTankNumPlayer1] = tank
	return tank, nil
}

// CompleteSpawn завершает появление, как настоящий жизненный цикл
func (s *stubLifecycle) CompleteSpawn(tank *types.TankEntity) {
	if tank != nil && tank.State == types.TankStateSpawning {
		tank.State = types.TankStateStopped
	}
}

func (s *stubLifecycle) SpawnPlayer2(uint) (*types.TankEntity, error) {
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
func (s *stubTankCommon) SetMaxLevel(*types.TankEntity)    {}
func (s *stubTankCommon) IsFrozen(*types.TankEntity) bool  { return false }

func (s *stubTankCommon) GetTankAnimationName(
	tank *types.TankEntity,
) string {
	return ""
}

type stubBulletUseCases struct {
	updateCalls int
}

func (s *stubBulletUseCases) ShootBullet(tank *types.TankEntity) (bool, error) {
	return false, nil
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

// stubSpawnSelection выдаёт спаунеры по кругу; blocked — все заняты
type stubSpawnSelection struct {
	spawners int
	next     int
	blocked  bool
}

func (s *stubSpawnSelection) SelectSpawner(
	session *session_entities.StageSessionEntity,
) (int, bool) {
	if s.blocked {
		return 0, false
	}
	index := s.next % s.spawners
	s.next++
	return index, true
}

func (s *stubSpawnSelection) GetSpawnersCount() int { return s.spawners }

// newTestLevel — уровень из волн в синтаксисе букв, без бонусной
// разметки, если в буквах нет строчных
func newTestLevel(maxActive uint, waves ...types.WaveSpec) *types.LevelEntity {
	explicit := false
	for _, wave := range waves {
		for _, tank := range wave.Tanks {
			explicit = explicit || tank.HasBonus
		}
	}
	return types.NewLevelEntity(1, "TEST", maxActive, 600, waves, explicit, nil)
}

// basicWave — волна из count обычных танков
func basicWave(count int, delay uint, start types.WaveStart) types.WaveSpec {
	return types.WaveSpec{
		Tanks:      make([]types.WaveTank, count),
		DelayTicks: delay,
		Start:      start,
	}
}

type stageTestEnv struct {
	selection *stubSpawnSelection
	lifecycle *stubLifecycle
	common    *stubTankCommon
	bullets   *stubBulletUseCases
	collision *stubCollisionUseCases
	hq        *stubHQUseCases
	session   *session_entities.StageSessionEntity
	bonuses   *game.BonusesRepository
	stage     *state_use_cases.StageUseCases
}

// newStageTestEnv — уровень из одной волны в 20 обычных танков
// с паузой enemyRespawnDelay и лимитом 5 активных врагов
func newStageTestEnv(enemyRespawnDelay uint) *stageTestEnv {
	return newStageTestEnvWithLevel(newTestLevel(
		5, basicWave(20, enemyRespawnDelay, types.WaveStart{}),
	))
}

func newStageTestEnvWithLevel(level *types.LevelEntity) *stageTestEnv {
	selection := &stubSpawnSelection{spawners: 3}
	lifecycle := &stubLifecycle{}
	common := &stubTankCommon{}
	bullets := &stubBulletUseCases{}
	collision := &stubCollisionUseCases{}
	hq := &stubHQUseCases{}
	session := session_entities.NewStageSessionEntity()
	session.SetUpLevel(level)
	bonuses := game.NewBonusesRepository()

	stage := state_use_cases.NewStageUseCases(
		lifecycle,
		use_cases.NewWaveUseCases(),
		selection,
		common,
		bullets,
		collision,
		hq,
		session,
		bonuses,
		nil, // mapUseCases: пути с ним защищены nil-проверками
		nil, // bonusUseCases: пути с ним защищены nil-проверками
	)

	return &stageTestEnv{
		selection: selection,
		lifecycle: lifecycle,
		common:    common,
		bullets:   bullets,
		collision: collision,
		hq:        hq,
		session:   session,
		bonuses:   bonuses,
		stage:     stage,
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
		nil, nil, nil, nil, nil, nil, nil, session, nil, nil, nil,
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
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
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
	if got := stage.PlacePlayerTank(types.TankRolePlayer1); got != nil {
		t.Errorf("PlacePlayerTank: %v", got)
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

// После спауна следующий враг ждёт паузу волны
func TestStageUseCases_TrySpawnEnemy_RespawnDelay(t *testing.T) {
	env := newStageTestEnv(2)
	env.lifecycle.nextEnemy = newEnemyFactory()

	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("the first enemy must spawn without a pause")
	}
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("enemy spawned before the pause")
	}
	env.stage.UpdateGameObjects(testDT)
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("enemy spawned in the middle of the pause")
	}
	env.stage.UpdateGameObjects(testDT)
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("enemy did not spawn after the pause")
	}

	if len(env.lifecycle.spawnCalls) != 2 {
		t.Fatalf("spawn calls %d, want 2", len(env.lifecycle.spawnCalls))
	}
	if call := env.lifecycle.spawnCalls[1]; call.spawnerIndex != 1 ||
		call.level != types.EnemyLevelBasic {
		t.Errorf("spawn arguments: %+v", call)
	}
	if got := env.session.GetRecentSpawners(); got != [2]int{1, 0} {
		t.Errorf("recent spawners %v, want [1 0]", got)
	}
}

// Лимит одновременно активных врагов берётся из уровня
func TestStageUseCases_TrySpawnEnemy_MaxActiveEnemiesCap(t *testing.T) {
	env := newStageTestEnv(1)
	env.lifecycle.nextEnemy = newEnemyFactory()

	for i := 0; i < 5; i++ {
		env.common.tanks = append(
			env.common.tanks,
			newTankInState(types.TankRoleEnemy, types.TankStateStopped),
		)
	}

	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("enemy spawned with a full limit")
	}
	if len(env.lifecycle.spawnCalls) != 0 {
		t.Fatal("lifecycle called with a full limit")
	}

	// Взорванные враги и игроки не считаются активными врагами
	env.common.tanks[0].State = types.TankStateExploded
	env.common.tanks = append(
		env.common.tanks,
		newTankInState(types.TankRolePlayer1, types.TankStateMoving),
	)
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("enemy did not spawn into a free slot")
	}
}

// Вдвоём лимит активных врагов выше на два
func TestStageUseCases_TrySpawnEnemy_CoopRaisesLimit(t *testing.T) {
	env := newStageTestEnv(0)
	env.session.SetPlayerCount(2)
	env.lifecycle.nextEnemy = newEnemyFactory()
	for i := 0; i < 6; i++ {
		env.common.tanks = append(
			env.common.tanks,
			newTankInState(types.TankRoleEnemy, types.TankStateStopped),
		)
	}

	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("two players: the 7th enemy must fit into the limit")
	}
}

// Все спаунеры заняты: спаун откладывается, сценарий не сдвигается
func TestStageUseCases_TrySpawnEnemy_AllSpawnersBlocked(t *testing.T) {
	env := newStageTestEnv(0)
	env.lifecycle.nextEnemy = newEnemyFactory()
	env.selection.blocked = true

	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("enemy spawned with every spawner blocked")
	}
	if got := env.session.GetSpawnedEnemies(); got != 0 {
		t.Errorf("spawned %d, want 0", got)
	}
}

// Неудавшийся спавн не сдвигает номер врага и не сбрасывает отсчёт
func TestStageUseCases_TrySpawnEnemy_FailedSpawnKeepsSchedule(t *testing.T) {
	env := newStageTestEnv(1)

	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("nextEnemy nil: spawn must return nil")
	}
	if len(env.lifecycle.spawnCalls) != 1 {
		t.Fatalf("spawn calls %d, want 1", len(env.lifecycle.spawnCalls))
	}
	if got := env.session.GetNextEnemyNumber(); got != 1 {
		t.Errorf("enemy number moved: %d", got)
	}

	// Отсчёт не сброшен: повторная попытка проходит без ожидания
	env.lifecycle.nextEnemy = newEnemyFactory()
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("retry failed")
	}
	if got := env.session.GetNextEnemyNumber(); got != 2 {
		t.Errorf("enemy number after spawn %d, want 2", got)
	}
}

// Без явной разметки бонусные враги идут по классической
// нумерации: 4, 9, 15
func TestStageUseCases_BonusEnemySequence(t *testing.T) {
	env := newStageTestEnv(0)
	env.lifecycle.nextEnemy = newEnemyFactory()

	bonusNumbers := map[uint]bool{4: true, 9: true, 15: true}

	for number := uint(1); number <= 20; number++ {
		tank := env.stage.TrySpawnEnemy()
		if tank == nil {
			t.Fatalf("enemy %d did not spawn", number)
		}
		if tank.GetWithBonus() != bonusNumbers[number] {
			t.Errorf(
				"enemy %d: withBonus=%v, want %v",
				number,
				tank.GetWithBonus(),
				bonusNumbers[number],
			)
		}
	}

	// Все 20 врагов заспавнены — дальше спавн невозможен
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Error("spawn after all enemies are out")
	}
}

// Явная разметка: бонус несут только отмеченные танки,
// уровень танка берётся из волны
func TestStageUseCases_ExplicitWaveTanks(t *testing.T) {
	env := newStageTestEnvWithLevel(newTestLevel(5, types.WaveSpec{
		Tanks: []types.WaveTank{
			{Level: types.EnemyLevelArmor},
			{Level: types.EnemyLevelFast, HasBonus: true},
		},
	}))
	env.lifecycle.nextEnemy = newEnemyFactory()

	first := env.stage.TrySpawnEnemy()
	second := env.stage.TrySpawnEnemy()
	if first == nil || second == nil {
		t.Fatal("wave tanks did not spawn")
	}
	if first.GetWithBonus() || !second.GetWithBonus() {
		t.Error("only the marked tank must carry a bonus")
	}
	if env.lifecycle.spawnCalls[0].level != types.EnemyLevelArmor ||
		env.lifecycle.spawnCalls[1].level != types.EnemyLevelFast {
		t.Errorf("spawn levels %+v", env.lifecycle.spawnCalls)
	}
}

// Волна с условием clear ждёт уничтожения всех врагов прошлых волн
func TestStageUseCases_WaveWaitsForClear(t *testing.T) {
	env := newStageTestEnvWithLevel(newTestLevel(
		5,
		basicWave(2, 0, types.WaveStart{}),
		basicWave(1, 0, types.WaveStart{Kind: types.WaveStartClear}),
	))
	env.lifecycle.nextEnemy = newEnemyFactory()

	env.stage.TrySpawnEnemy()
	env.stage.TrySpawnEnemy()
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("clear wave started with enemies alive")
	}

	env.session.IncrementDestroyedEnemies()
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("clear wave started with one enemy alive")
	}

	env.session.IncrementDestroyedEnemies()
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("clear wave did not start after the field was cleared")
	}
}

// Волна left<=N стартует, когда живых врагов не больше N
func TestStageUseCases_WaveWaitsForLeft(t *testing.T) {
	env := newStageTestEnvWithLevel(newTestLevel(
		5,
		basicWave(3, 0, types.WaveStart{}),
		basicWave(1, 0, types.WaveStart{Kind: types.WaveStartLeft, Left: 1}),
	))
	env.lifecycle.nextEnemy = newEnemyFactory()

	for i := 0; i < 3; i++ {
		env.stage.TrySpawnEnemy()
	}
	env.session.IncrementDestroyedEnemies()
	if got := env.stage.TrySpawnEnemy(); got != nil {
		t.Fatal("left<=1 wave started with two enemies alive")
	}

	env.session.IncrementDestroyedEnemies()
	if got := env.stage.TrySpawnEnemy(); got == nil {
		t.Fatal("left<=1 wave did not start with one enemy alive")
	}
}

// Начальный спавн: по врагу на спаунер по порядку, не больше
// лимита активных
func TestStageUseCases_SpawnInitialEnemyTanks(t *testing.T) {
	env := newStageTestEnvWithLevel(newTestLevel(
		2, basicWave(5, 90, types.WaveStart{}),
	))
	env.lifecycle.nextEnemy = newEnemyFactory()

	spawned := env.stage.SpawnInitialEnemyTanks()

	if len(spawned) != 2 {
		t.Fatalf("spawned %d, want 2 (max_active)", len(spawned))
	}
	for i, call := range env.lifecycle.spawnCalls {
		if call.spawnerIndex != i {
			t.Errorf("enemy %d at spawner %d", i, call.spawnerIndex)
		}
	}
	if got := env.session.GetNextEnemyNumber(); got != 3 {
		t.Errorf("next enemy number %d, want 3", got)
	}
	if env.session.CanSpawnNextEnemy() {
		t.Error("the wave pause must start after the initial spawn")
	}
}

// Итог уровня: исход, потерянные жизни и время без пауз
func TestStageUseCases_GetStageResult(t *testing.T) {
	env := newStageTestEnvWithLevel(newTestLevel(
		5, basicWave(1, 0, types.WaveStart{}),
	))

	env.stage.UpdateGameObjects(testDT)
	env.stage.PauseStageState()
	env.stage.UpdateGameObjects(testDT)
	env.stage.ResumeStageState()
	env.stage.UpdateGameObjects(testDT)
	env.session.DecrementPlayerLives(types.PlayerTankNumPlayer1)
	env.session.IncrementDestroyedEnemies()

	result := env.stage.GetStageResult()
	if !result.Won || result.LivesLost != 1 || result.ElapsedTicks != 2 {
		t.Errorf("result %+v, want won, 1 life lost, 2 ticks", result)
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

// На старте уровня танк игрока сразу на карте, без анимации появления
func TestStageUseCases_PlacePlayerTank(t *testing.T) {
	env := newStageTestEnv(1)
	playerTank := newTankInState(
		types.TankRolePlayer1,
		types.TankStateSpawning,
	)
	env.lifecycle.nextPlayer1 = func() *types.TankEntity { return playerTank }

	if got := env.stage.PlacePlayerTank(types.TankRolePlayer1); got != playerTank {
		t.Fatalf("игрок 1: %v", got)
	}
	if !playerTank.IsActive() {
		t.Errorf("состояние %v, ожидался активный танк", playerTank.State)
	}

	// Побеждённый игрок на карту не ставится
	env.session.SetPlayerLives(types.PlayerTankNumPlayer1, 0)
	if got := env.stage.PlacePlayerTank(types.TankRolePlayer1); got != nil {
		t.Errorf("побеждённый игрок: %v", got)
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
	if got := env.session.GetPlayerDeaths(); got != 0 {
		t.Errorf("blocked respawn counted as a death: %d", got)
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

func TestStageUseCases_CarryOver(t *testing.T) {
	env := newStageTestEnv(1)
	p1 := types.PlayerTankNumPlayer1
	playerTank := newTankInState(types.TankRolePlayer1, types.TankStateStopped)
	playerTank.SetSpecs(types.NewSpecsEntity(2, 32, true, 150, 1))
	env.lifecycle.players[p1] = playerTank
	env.session.SetPlayerLives(p1, 5)

	env.stage.SaveCarryOver()
	env.session.SetCarryOver(true)
	env.session.Reset()

	if got := env.session.GetPlayerLives(p1); got != 5 {
		t.Errorf("lives %d, want 5", got)
	}
	env.lifecycle.nextPlayer1 = func() *types.TankEntity {
		return newTankInState(types.TankRolePlayer1, types.TankStateSpawning)
	}
	env.stage.PlacePlayerTank(types.TankRolePlayer1)
	if env.lifecycle.player1Level != 2 {
		t.Errorf("placed with tier %d, want 2", env.lifecycle.player1Level)
	}
	// Возрождение — всегда с нулевой прокачкой
	env.stage.SpawnPlayerTank(types.TankRolePlayer1)
	if env.lifecycle.player1Level != 0 {
		t.Errorf("respawned with tier %d, want 0", env.lifecycle.player1Level)
	}
	if !env.stage.GetStageResult().CarriedOver {
		t.Error("result not marked as carried over")
	}
}

func TestStageUseCases_SaveCarryOverDestroyedTank(t *testing.T) {
	env := newStageTestEnv(1)
	p1 := types.PlayerTankNumPlayer1
	playerTank := newTankInState(types.TankRolePlayer1, types.TankStateExploded)
	playerTank.SetSpecs(types.NewSpecsEntity(3, 40, true, 150, 2))
	env.lifecycle.players[p1] = playerTank
	env.session.SetPlayerLives(p1, 5)

	env.stage.SaveCarryOver()
	env.session.SetCarryOver(true)
	env.session.Reset()

	// Несписанная жизнь уходит, прокачка теряется
	if got := env.session.GetPlayerLives(p1); got != 4 {
		t.Errorf("lives %d, want 4", got)
	}
	if got := env.session.GetStartTier(p1); got != 0 {
		t.Errorf("tier %d, want 0", got)
	}
}
