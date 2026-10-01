package tank_use_cases_test

import (
	"errors"
	"image/color"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

// forcedLevelSpecs делегирует реальным спецификациям
type forcedLevelSpecs struct {
	real *use_cases.SpecsUseCases
}

func (s *forcedLevelSpecs) GetTankSpecs(
	isEnemy bool,
	level uint,
) *types.SpecsEntity {
	return s.real.GetTankSpecs(isEnemy, level)
}

type stubRenderUseCases struct {
	spawnFinished     bool
	explosionFinished bool
	updatedAnimations []*types.TankEntity
}

func (s *stubRenderUseCases) IsTankSpawnAnimationFinished(
	tank *types.TankEntity,
) bool {
	return s.spawnFinished
}

func (s *stubRenderUseCases) IsTankExplosionAnimationFinished(
	tank *types.TankEntity,
) bool {
	return s.explosionFinished
}

func (s *stubRenderUseCases) UpdateTankAnimation(tank *types.TankEntity) {
	s.updatedAnimations = append(s.updatedAnimations, tank)
}

func (s *stubRenderUseCases) SyncTankAnimationWithState(
	tank *types.TankEntity,
) {
}

func (s *stubRenderUseCases) UpdateBlink(blinkObjects []types.IBlink) {}

func (s *stubRenderUseCases) IsTankVisible(tank *types.TankEntity) bool {
	return true
}

func (s *stubRenderUseCases) IsTankBlinking(tank *types.TankEntity) bool {
	return false
}

func (s *stubRenderUseCases) BonusPulse(*types.TankEntity) (float64, bool) {
	return 0, false
}

func (s *stubRenderUseCases) TankHealthTint(
	tank *types.TankEntity,
) (color.NRGBA, bool) {
	return color.NRGBA{}, false
}

// stubSpawnCollisionService фиксирует проверенные позиции и
// позволяет объявить спавнер заблокированным
type stubSpawnCollisionService struct {
	blocked bool
	checked []types.Position
}

func (s *stubSpawnCollisionService) IsSpawnerBlocked(
	position types.Position,
	size types.Size,
	tanks []*types.TankEntity,
) bool {
	s.checked = append(s.checked, position)
	return s.blocked
}

type lifecycleTestEnv struct {
	tanksRepo      *game.TanksRepository
	animations     *game.AnimationsRepository
	tileService    *testutil.FakeTileService
	render         *stubRenderUseCases
	spawnCollision *stubSpawnCollisionService
	specs          *forcedLevelSpecs
	effects        *testutil.FakeVisualEffectsUseCases
	lifecycle      *tank_use_cases.TankLifecycleUseCases
}

// Спавнеры в тайлах, позиция танка = спавнер * baseSize (16px)
var (
	testEnemySpawners = []types.Position{
		{X: 2, Y: 0},
		{X: 6, Y: 0},
		{X: 12, Y: 0},
	}
	testPlayer1Spawner = types.Position{X: 4, Y: 12}
	testPlayer2Spawner = types.Position{X: 8, Y: 12}
)

func newLifecycleTestEnv() *lifecycleTestEnv {
	tanksRepo := game.NewTanksRepository()
	animations := game.NewAnimationsRepository()
	tileService := &testutil.FakeTileService{}
	tilesUC := use_cases.NewTilesUseCasesWithAnimations(
		nil, // реестр тайлсетов не нужен для анимаций спавна и взрыва
		types.TilesetTypePlayer,
		animations,
		tileService,
		nil,
	)
	specs := &forcedLevelSpecs{real: use_cases.NewSpecsUseCases()}
	render := &stubRenderUseCases{}
	common := tank_use_cases.NewTankCommonUseCases(
		services.NewTankBrakingService(),
		render,
		tanksRepo,
		specs,
		use_cases.NewMapUseCases(nil),
		session_entities.NewStageSessionEntity(),
	)
	spawnCollision := &stubSpawnCollisionService{}
	effects := &testutil.FakeVisualEffectsUseCases{}
	lifecycle := tank_use_cases.NewTankLifecycleUseCases(
		tilesUC,
		render,
		common,
		tanksRepo,
		spawnCollision,
		specs,
		effects,
		types.SpawnLayout{
			EnemySpawners:  testEnemySpawners,
			Player1Spawner: testPlayer1Spawner,
			Player2Spawner: testPlayer2Spawner,
			BaseSize:       types.Size{Width: 16, Height: 16},
		},
	)

	return &lifecycleTestEnv{
		tanksRepo:      tanksRepo,
		animations:     animations,
		tileService:    tileService,
		render:         render,
		spawnCollision: spawnCollision,
		specs:          specs,
		effects:        effects,
		lifecycle:      lifecycle,
	}
}

func assertAnimatingImage(t *testing.T, tank *types.TankEntity) {
	t.Helper()
	animation, ok := tank.Image.(*image_providers.AnimationProvider)
	if !ok {
		t.Fatalf("Image не AnimationProvider: %T", tank.Image)
	}
	if !animation.IsAnimating {
		t.Error("анимация спавна не запущена")
	}
}

func TestTankLifecycleUseCases_SpawnPlayer1(t *testing.T) {
	env := newLifecycleTestEnv()

	tank, err := env.lifecycle.SpawnPlayer1(0)
	if err != nil || tank == nil {
		t.Fatalf("спавн игрока 1: tank=%v err=%v", tank, err)
	}

	want := types.Position{X: 4 * 16, Y: 12 * 16}
	if tank.Position != want {
		t.Errorf("позиция %v, ожидалась %v", tank.Position, want)
	}
	if tank.GetRole() != types.TankRolePlayer1 {
		t.Errorf("роль %q", tank.GetRole())
	}
	if tank.State != types.TankStateSpawning {
		t.Errorf("состояние %v, ожидалось Spawning", tank.State)
	}
	if tank.Altitude != types.SURFACE {
		t.Errorf("высота %v, ожидалась SURFACE", tank.Altitude)
	}
	if tank.Size != (types.Size{Width: 16, Height: 16}) {
		t.Errorf("размер %v", tank.Size)
	}
	if got := tank.GetHitPoints(); got != 1 {
		t.Errorf("хитпоинты %d, ожидалось 1", got)
	}
	if got := tank.GetSpecs().GetLevel(); got != 0 {
		t.Errorf("уровень %d, ожидался 0", got)
	}
	if env.tanksRepo.GetPlayer(types.PlayerTankNumPlayer1) != tank {
		t.Error("танк не зарегистрирован как игрок 1")
	}
	assertAnimatingImage(t, tank)
	if got := len(env.animations.GetAllAnimations()); got != 1 {
		t.Errorf("анимаций в репозитории %d, ожидалась 1", got)
	}
}

func TestTankLifecycleUseCases_SpawnPlayer2(t *testing.T) {
	env := newLifecycleTestEnv()

	tank, err := env.lifecycle.SpawnPlayer2(0)
	if err != nil || tank == nil {
		t.Fatalf("спавн игрока 2: tank=%v err=%v", tank, err)
	}

	want := types.Position{X: 8 * 16, Y: 12 * 16}
	if tank.Position != want {
		t.Errorf("позиция %v, ожидалась %v", tank.Position, want)
	}
	if tank.GetRole() != types.TankRolePlayer2 {
		t.Errorf("роль %q", tank.GetRole())
	}
	if env.tanksRepo.GetPlayer(types.PlayerTankNumPlayer2) != tank {
		t.Error("танк не зарегистрирован как игрок 2")
	}
}

// Заблокированный спавнер: игрок не создаётся, ошибки нет
func TestTankLifecycleUseCases_SpawnPlayerBlockedSpawner(t *testing.T) {
	env := newLifecycleTestEnv()
	env.spawnCollision.blocked = true

	tank, err := env.lifecycle.SpawnPlayer1(0)
	if tank != nil || err != nil {
		t.Fatalf("ожидалось nil, nil; получено tank=%v err=%v", tank, err)
	}
	if env.tanksRepo.HasPlayer(types.PlayerTankNumPlayer1) {
		t.Error("игрок зарегистрирован при заблокированном спавнере")
	}
	if len(env.spawnCollision.checked) != 1 ||
		env.spawnCollision.checked[0] != testPlayer1Spawner {
		t.Errorf("проверена не та позиция: %v", env.spawnCollision.checked)
	}
}

func TestTankLifecycleUseCases_SpawnEnemy_ByIndex(t *testing.T) {
	env := newLifecycleTestEnv()

	tank, err := env.lifecycle.SpawnEnemy(1, types.EnemyLevelBasic)
	if err != nil || tank == nil {
		t.Fatalf("spawn: tank=%v err=%v", tank, err)
	}

	want := types.Position{X: 6 * 16, Y: 0}
	if tank.Position != want {
		t.Errorf("position %v, want %v", tank.Position, want)
	}
	if !tank.IsEnemy() {
		t.Errorf("role %q, want enemy", tank.GetRole())
	}
	if tank.Direction != types.DirectionUp {
		t.Errorf("direction %v, want Up", tank.Direction)
	}
	if tank.State != types.TankStateSpawning {
		t.Errorf("state %v", tank.State)
	}
	if got := tank.GetSpecs().GetLevel(); got != 0 {
		t.Errorf("level %d, want 0", got)
	}
	if got := tank.GetHitPoints(); got != 1 {
		t.Errorf("hit points %d, want 1", got)
	}

	enemies := env.tanksRepo.GetAllEnemies()
	if len(enemies) != 1 || enemies[0] != tank {
		t.Errorf("enemy not added to the repository: %v", enemies)
	}
}

// Тяжёлый танк (враг 3 уровня) получает 4 хитпоинта
func TestTankLifecycleUseCases_SpawnEnemyHeavyTankHitPoints(t *testing.T) {
	env := newLifecycleTestEnv()

	tank, err := env.lifecycle.SpawnEnemy(0, types.EnemyLevelArmor)
	if err != nil || tank == nil {
		t.Fatalf("spawn: tank=%v err=%v", tank, err)
	}
	if got := tank.GetSpecs().GetLevel(); got != 3 {
		t.Errorf("level %d, want 3", got)
	}
	if got := tank.GetHitPoints(); got != 4 {
		t.Errorf("hit points %d, want 4", got)
	}
}

// Занятость спаунера проверяет вызывающий: SpawnEnemy её не смотрит
func TestTankLifecycleUseCases_SpawnEnemyIgnoresBlocking(t *testing.T) {
	env := newLifecycleTestEnv()
	env.spawnCollision.blocked = true

	tank, err := env.lifecycle.SpawnEnemy(0, types.EnemyLevelBasic)
	if err != nil || tank == nil {
		t.Fatalf("spawn: tank=%v err=%v", tank, err)
	}
}

func TestTankLifecycleUseCases_SpawnEnemyIndexOutOfRange(t *testing.T) {
	env := newLifecycleTestEnv()

	for _, index := range []int{-1, len(testEnemySpawners)} {
		if _, err := env.lifecycle.SpawnEnemy(index, 0); err == nil {
			t.Errorf("index %d: expected an out of range error", index)
		}
	}
}

func TestTankLifecycleUseCases_Explode(t *testing.T) {
	env := newLifecycleTestEnv()
	tank, err := env.lifecycle.SpawnPlayer1(0)
	if err != nil || tank == nil {
		t.Fatalf("спавн: tank=%v err=%v", tank, err)
	}
	spawnImage := tank.Image

	if err := env.lifecycle.Explode(tank); err != nil {
		t.Fatalf("взрыв: %v", err)
	}
	if tank.State != types.TankStateExploding {
		t.Errorf("состояние %v, ожидалось Exploding", tank.State)
	}
	if tank.Altitude != types.AIR {
		t.Errorf("высота %v, ожидалась AIR", tank.Altitude)
	}
	if tank.Image == spawnImage {
		t.Error("изображение не заменено анимацией взрыва")
	}
	assertAnimatingImage(t, tank)
	if kinds := env.effects.Kinds(); len(kinds) != 1 ||
		kinds[0] != types.VisualEventTankExplosion ||
		env.effects.Events[0].Tank != tank {
		t.Errorf("события эффектов %v, ожидался взрыв танка", kinds)
	}

	// Создавалась и анимация спавна, и анимация взрыва
	if len(env.tileService.Created) != 2 ||
		env.tileService.Created[0] != "spawner/spawner" ||
		env.tileService.Created[1] != "explosion/explosion_tank" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}
}

// Ошибка тайл-сервиса прерывает и спавн, и взрыв
func TestTankLifecycleUseCases_TileServiceError(t *testing.T) {
	env := newLifecycleTestEnv()
	env.tileService.Err = errors.New("tileset missing")

	if tank, err := env.lifecycle.SpawnPlayer1(0); err == nil || tank != nil {
		t.Errorf("ожидалась ошибка спавна, tank=%v err=%v", tank, err)
	}

	tankValue := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	tank := &tankValue
	tank.State = types.TankStateStopped
	if err := env.lifecycle.Explode(tank); err == nil {
		t.Error("ожидалась ошибка взрыва")
	}
	if tank.State != types.TankStateStopped {
		t.Errorf("состояние изменилось при ошибке: %v", tank.State)
	}
}

// Завершение анимации спавна переводит танк в Stopped
func TestTankLifecycleUseCases_UpdateLifecycle_SpawnToStopped(t *testing.T) {
	env := newLifecycleTestEnv()
	tank, err := env.lifecycle.SpawnPlayer1(0)
	if err != nil || tank == nil {
		t.Fatalf("спавн: tank=%v err=%v", tank, err)
	}

	// Анимация ещё идёт — состояние не меняется
	if err := env.lifecycle.UpdateAllTanksLifecycle(); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.State != types.TankStateSpawning {
		t.Fatalf("состояние %v, ожидалось Spawning", tank.State)
	}

	env.render.spawnFinished = true
	if err := env.lifecycle.UpdateAllTanksLifecycle(); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.State != types.TankStateStopped {
		t.Errorf("состояние %v, ожидалось Stopped", tank.State)
	}
	if len(env.render.updatedAnimations) != 1 ||
		env.render.updatedAnimations[0] != tank {
		t.Errorf(
			"анимация танка не обновлена: %v",
			env.render.updatedAnimations,
		)
	}
}

// Завершение анимации взрыва переводит танк в Exploded
func TestTankLifecycleUseCases_UpdateLifecycle_ExplodingToExploded(
	t *testing.T,
) {
	env := newLifecycleTestEnv()
	tank, err := env.lifecycle.SpawnEnemy(0, types.EnemyLevelBasic)
	if err != nil || tank == nil {
		t.Fatalf("спавн: tank=%v err=%v", tank, err)
	}
	if err := env.lifecycle.Explode(tank); err != nil {
		t.Fatalf("взрыв: %v", err)
	}

	if err := env.lifecycle.UpdateAllTanksLifecycle(); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.State != types.TankStateExploding {
		t.Fatalf("состояние %v, ожидалось Exploding", tank.State)
	}

	env.render.explosionFinished = true
	if err := env.lifecycle.UpdateAllTanksLifecycle(); err != nil {
		t.Fatalf("обновление: %v", err)
	}
	if tank.State != types.TankStateExploded {
		t.Errorf("состояние %v, ожидалось Exploded", tank.State)
	}
}

func TestTankLifecycleUseCases_GetSetPlayerTank(t *testing.T) {
	env := newLifecycleTestEnv()

	if got := env.lifecycle.GetPlayerTank(types.PlayerTankNumPlayer1); got != nil {
		t.Errorf("ожидался nil, получен %v", got)
	}

	tankValue := types.NewDefaultTankEntity(
		types.TankRolePlayer1,
		types.DirectionUp,
	)
	tank := &tankValue
	env.lifecycle.SetPlayerTank(types.PlayerTankNumPlayer1, tank)
	if got := env.lifecycle.GetPlayerTank(types.PlayerTankNumPlayer1); got != tank {
		t.Errorf("танк не сохранён: %v", got)
	}
}

// CompleteSpawn ставит появляющийся танк на карту сразу;
// активный танк не трогает
func TestTankLifecycleUseCases_CompleteSpawn(t *testing.T) {
	env := newLifecycleTestEnv()
	tank, err := env.lifecycle.SpawnPlayer1(0)
	if err != nil || tank == nil {
		t.Fatalf("спавн: tank=%v err=%v", tank, err)
	}

	env.lifecycle.CompleteSpawn(tank)

	if tank.State != types.TankStateStopped {
		t.Errorf("состояние %v, ожидалось Stopped", tank.State)
	}
	if got := len(env.render.updatedAnimations); got != 1 {
		t.Errorf("анимация танка обновлена %d раз, ожидался 1", got)
	}
	if !tank.HasShield() {
		t.Error("появившийся игрок без неуязвимости")
	}

	tank.State = types.TankStateMoving
	env.lifecycle.CompleteSpawn(tank)
	if tank.State != types.TankStateMoving {
		t.Error("активный танк не должен меняться")
	}
}

// После появления неуязвим только игрок, как в оригинале
func TestTankLifecycleUseCases_SpawnShieldOnlyForPlayer(t *testing.T) {
	env := newLifecycleTestEnv()
	env.render.spawnFinished = true
	player, _ := env.lifecycle.SpawnPlayer1(0)
	enemy, err := env.lifecycle.SpawnEnemy(0, types.EnemyLevelBasic)
	if err != nil || player == nil || enemy == nil {
		t.Fatalf("спавн: player=%v enemy=%v err=%v", player, enemy, err)
	}

	_ = env.lifecycle.UpdateAllTanksLifecycle()

	if !player.IsActive() || !player.HasShield() {
		t.Error("возродившийся игрок должен быть неуязвим")
	}
	if !enemy.IsActive() || enemy.HasShield() {
		t.Error("враг после появления не должен быть неуязвим")
	}
}
