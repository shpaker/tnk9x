package tank_use_cases_test

import (
	"errors"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/use_cases"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

// forcedLevelSpecs делегирует реальным спецификациям, но позволяет
// зафиксировать уровень врага вместо случайного выбора
type forcedLevelSpecs struct {
	real        *use_cases.SpecsUseCases
	forcedLevel *uint
}

func (s *forcedLevelSpecs) GetTankSpecs(
	isEnemy bool,
	level uint,
) *types.SpecsEntity {
	return s.real.GetTankSpecs(isEnemy, level)
}

func (s *forcedLevelSpecs) GetEnemyLevelByRemainingCount(
	remainingEnemies uint,
) uint {
	if s.forcedLevel != nil {
		return *s.forcedLevel
	}
	return s.real.GetEnemyLevelByRemainingCount(remainingEnemies)
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
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
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
	)
	spawnCollision := &stubSpawnCollisionService{}
	lifecycle := tank_use_cases.NewTankLifecycleUseCases(
		tilesUC,
		render,
		common,
		tanksRepo,
		spawnCollision,
		specs,
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

	tank, err := env.lifecycle.SpawnPlayer1()
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

	tank, err := env.lifecycle.SpawnPlayer2()
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

	tank, err := env.lifecycle.SpawnPlayer1()
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

func TestTankLifecycleUseCases_SpawnEnemyWithLevel_ByIndex(t *testing.T) {
	env := newLifecycleTestEnv()
	index := 1

	tank, err := env.lifecycle.SpawnEnemyWithLevel(&index, false, 20)
	if err != nil || tank == nil {
		t.Fatalf("спавн врага: tank=%v err=%v", tank, err)
	}

	want := types.Position{X: 6 * 16, Y: 0}
	if tank.Position != want {
		t.Errorf("позиция %v, ожидалась %v", tank.Position, want)
	}
	if !tank.IsEnemy() {
		t.Errorf("роль %q, ожидался враг", tank.GetRole())
	}
	if tank.Direction != types.DirectionUp {
		t.Errorf("направление %v, ожидалось Up", tank.Direction)
	}
	if tank.State != types.TankStateSpawning {
		t.Errorf("состояние %v", tank.State)
	}
	// Оставшихся 20 -> уровень 0, обычный танк с 1 хитпоинтом
	if got := tank.GetSpecs().GetLevel(); got != 0 {
		t.Errorf("уровень %d, ожидался 0", got)
	}
	if got := tank.GetHitPoints(); got != 1 {
		t.Errorf("хитпоинты %d, ожидалось 1", got)
	}

	enemies := env.tanksRepo.GetAllEnemies()
	if len(enemies) != 1 || enemies[0] != tank {
		t.Errorf("враг не добавлен в репозиторий: %v", enemies)
	}
}

// Тяжёлый танк (враг 3 уровня) получает 4 хитпоинта
func TestTankLifecycleUseCases_SpawnEnemyHeavyTankHitPoints(t *testing.T) {
	env := newLifecycleTestEnv()
	level := uint(3)
	env.specs.forcedLevel = &level
	index := 0

	tank, err := env.lifecycle.SpawnEnemyWithLevel(&index, false, 1)
	if err != nil || tank == nil {
		t.Fatalf("спавн врага: tank=%v err=%v", tank, err)
	}
	if got := tank.GetSpecs().GetLevel(); got != 3 {
		t.Errorf("уровень %d, ожидался 3", got)
	}
	if got := tank.GetHitPoints(); got != 4 {
		t.Errorf("хитпоинты %d, ожидалось 4", got)
	}
}

func TestTankLifecycleUseCases_SpawnEnemyBlockedSpawner(t *testing.T) {
	env := newLifecycleTestEnv()
	env.spawnCollision.blocked = true
	index := 0

	// Без ignoreRespawnDelay блокировка отменяет спавн без ошибки
	tank, err := env.lifecycle.SpawnEnemyWithLevel(&index, false, 20)
	if tank != nil || err != nil {
		t.Fatalf("ожидалось nil, nil; получено tank=%v err=%v", tank, err)
	}
	if got := len(env.tanksRepo.GetAllEnemies()); got != 0 {
		t.Errorf("враг добавлен при блокировке: %d", got)
	}

	// С ignoreRespawnDelay спавн проходит несмотря на блокировку
	tank, err = env.lifecycle.SpawnEnemyWithLevel(&index, true, 20)
	if err != nil || tank == nil {
		t.Fatalf("ожидался спавн: tank=%v err=%v", tank, err)
	}
}

func TestTankLifecycleUseCases_SpawnEnemyIndexOutOfRange(t *testing.T) {
	env := newLifecycleTestEnv()
	index := len(testEnemySpawners)

	if _, err := env.lifecycle.SpawnEnemyWithLevel(&index, false, 20); err == nil {
		t.Error("ожидалась ошибка для индекса вне диапазона")
	}
}

// Без индекса спавнер выбирается случайно из настроенных
func TestTankLifecycleUseCases_SpawnEnemyRandomSpawner(t *testing.T) {
	env := newLifecycleTestEnv()

	allowed := make(map[types.Position]bool, len(testEnemySpawners))
	for _, spawner := range testEnemySpawners {
		allowed[types.Position{X: spawner.X * 16, Y: spawner.Y * 16}] = true
	}

	for i := 0; i < 20; i++ {
		tank, err := env.lifecycle.SpawnEnemyWithLevel(nil, false, 20)
		if err != nil || tank == nil {
			t.Fatalf("спавн врага: tank=%v err=%v", tank, err)
		}
		if !allowed[tank.Position] {
			t.Fatalf("недопустимая позиция спавна: %v", tank.Position)
		}
	}
}

// Начальный спавн: три врага 0 уровня на всех спавнерах,
// блокировка спавнера игнорируется
func TestTankLifecycleUseCases_OnStageSetUpEnemiesSpawn(t *testing.T) {
	env := newLifecycleTestEnv()
	env.spawnCollision.blocked = true

	spawned, err := env.lifecycle.OnStageSetUpEnemiesSpawn()
	if err != nil {
		t.Fatalf("начальный спавн: %v", err)
	}

	for i, tank := range spawned {
		if tank == nil {
			t.Fatalf("враг %d не создан", i)
		}
		want := types.Position{
			X: testEnemySpawners[i].X * 16,
			Y: testEnemySpawners[i].Y * 16,
		}
		if tank.Position != want {
			t.Errorf(
				"враг %d: позиция %v, ожидалась %v",
				i,
				tank.Position,
				want,
			)
		}
		if got := tank.GetSpecs().GetLevel(); got != 0 {
			t.Errorf("враг %d: уровень %d, ожидался 0", i, got)
		}
	}
	if got := len(env.tanksRepo.GetAllEnemies()); got != 3 {
		t.Errorf("врагов в репозитории %d, ожидалось 3", got)
	}
}

// Без настроенных спавнеров врагов случайный спавн невозможен,
// а начальный спавн молча возвращает пустой результат
func TestTankLifecycleUseCases_NoEnemySpawners(t *testing.T) {
	env := newLifecycleTestEnv()
	lifecycle := tank_use_cases.NewTankLifecycleUseCases(
		nil,
		env.render,
		nil,
		env.tanksRepo,
		env.spawnCollision,
		env.specs,
		types.SpawnLayout{BaseSize: types.Size{Width: 16, Height: 16}},
	)

	if _, err := lifecycle.SpawnEnemyWithLevel(nil, false, 20); err == nil {
		t.Error("SpawnEnemyWithLevel: ожидалась ошибка без спавнеров")
	}

	spawned, err := lifecycle.OnStageSetUpEnemiesSpawn()
	if err != nil {
		t.Fatalf("начальный спавн: %v", err)
	}
	for i, tank := range spawned {
		if tank != nil {
			t.Errorf("враг %d создан без спавнеров", i)
		}
	}
}

func TestTankLifecycleUseCases_Explode(t *testing.T) {
	env := newLifecycleTestEnv()
	tank, err := env.lifecycle.SpawnPlayer1()
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

	if tank, err := env.lifecycle.SpawnPlayer1(); err == nil || tank != nil {
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
	tank, err := env.lifecycle.SpawnPlayer1()
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
	index := 0
	tank, err := env.lifecycle.SpawnEnemyWithLevel(&index, false, 20)
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
