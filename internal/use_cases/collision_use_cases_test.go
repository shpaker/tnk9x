package use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services"
	"github.com/shpaker/tnk9x/internal/services/collision_services"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

const testDT = 1.0 / 60.0

type stubImageProvider struct{}

func (s *stubImageProvider) GetImageID() (string, error) { return "stub", nil }

// stubRenderUseCases — no-op рендер для сценариев движения и стрельбы
type stubRenderUseCases struct{}

func (s *stubRenderUseCases) IsTankSpawnAnimationFinished(
	tank *types.TankEntity,
) bool {
	return false
}

func (s *stubRenderUseCases) IsTankExplosionAnimationFinished(
	tank *types.TankEntity,
) bool {
	return false
}

func (s *stubRenderUseCases) UpdateTankAnimation(tank *types.TankEntity) {}

func (s *stubRenderUseCases) SyncTankAnimationWithState(
	tank *types.TankEntity,
) {
}

func (s *stubRenderUseCases) UpdateBlink(blinkObjects []types.IBlink) {}

type stubHQUseCases struct{}

func (s *stubHQUseCases) GetHQ() *types.HQEntity           { return nil }
func (s *stubHQUseCases) Explode(hq *types.HQEntity) error { return nil }
func (s *stubHQUseCases) IsExplosionFinished(hq *types.HQEntity) {
}
func (s *stubHQUseCases) IsDestroyed() bool { return false }

type stubTankLifecycle struct {
	exploded []*types.TankEntity
}

func (s *stubTankLifecycle) OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error) {
	return [3]*types.TankEntity{}, nil
}

func (s *stubTankLifecycle) SpawnEnemyWithLevel(
	index *int,
	ignoreRespawnDelay bool,
	remainingEnemies uint,
) (*types.TankEntity, error) {
	return nil, nil
}

func (s *stubTankLifecycle) SpawnPlayer1() (*types.TankEntity, error) { return nil, nil }

func (s *stubTankLifecycle) SpawnPlayer2() (*types.TankEntity, error) { return nil, nil }

func (s *stubTankLifecycle) GetPlayerTank(
	num types.PlayerTankNum,
) *types.TankEntity {
	return nil
}

func (s *stubTankLifecycle) SetPlayerTank(
	num types.PlayerTankNum,
	tank *types.TankEntity,
) {
}

func (s *stubTankLifecycle) Explode(tank *types.TankEntity) error {
	s.exploded = append(s.exploded, tank)
	tank.State = types.TankStateExploding
	return nil
}

func (s *stubTankLifecycle) UpdateAllTanksLifecycle() error { return nil }

type stubMapUseCases struct {
	mapEntity *types.MapEntity
}

func (s *stubMapUseCases) GetBlocks() types.MapBlocks { return s.mapEntity.GetBlocks() }

func (s *stubMapUseCases) RemoveBlock(block *types.BlockEntity) error {
	return s.mapEntity.RemoveBlock(block)
}

func (s *stubMapUseCases) GetSizePx() types.Size { return s.mapEntity.GetSizePx() }

func (s *stubMapUseCases) GetRandomBonusSpawnPosition() types.Position {
	return types.Position{}
}

func (s *stubMapUseCases) IsIceAt(_ types.Position) bool { return false }

type collisionTestEnv struct {
	tanksRepo   *game.TanksRepository
	bulletsRepo *game.BulletsRepository
	mapEntity   *types.MapEntity
	bulletUC    *use_cases.BulletUseCases
	tankCommon  *tank_use_cases.TankCommonUseCases
	tankActions *tank_use_cases.TankActionsUseCases
	collision   *use_cases.CollisionUseCases
	entities    *collision_services.EntitiesCollisionService
	lifecycle   *stubTankLifecycle
	specsUC     *use_cases.SpecsUseCases
}

func newCollisionTestEnv(blocks types.MapBlocks) *collisionTestEnv {
	tanksRepo := game.NewTanksRepository()
	bulletsRepo := game.NewBulletsRepository()
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		blocks,
		nil,
	)
	mapUC := &stubMapUseCases{mapEntity: mapEntity}

	entities := collision_services.NewEntitiesCollisionService()
	wall := collision_services.NewWallCollisionService(entities)
	boundary := collision_services.NewBoundaryCollisionService(
		types.Size{Width: 208, Height: 208},
	)
	bulletSvc := collision_services.NewBulletCollisionService(8, entities)
	braking := services.NewTankBrakingService()
	specsUC := use_cases.NewSpecsUseCases()

	render := &stubRenderUseCases{}
	soundUC := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())

	bulletUC := use_cases.NewBulletUseCases(bulletsRepo, nil, 16)
	tankCommon := tank_use_cases.NewTankCommonUseCases(
		braking,
		render,
		tanksRepo,
		specsUC,
		mapUC,
	)
	tankActions := tank_use_cases.NewTankActionsUseCases(
		braking,
		bulletUC,
		tankCommon,
		render,
		mapUC,
		soundUC,
	)
	lifecycle := &stubTankLifecycle{}

	collision := use_cases.NewCollisionUseCases(
		bulletUC,
		tankActions,
		mapUC,
		tankCommon,
		lifecycle,
		boundary,
		wall,
		bulletSvc,
		entities,
		collision_services.NewSpawnCollisionService(entities),
		&stubHQUseCases{},
		nil,
		game.NewBonusesRepository(),
		soundUC,
	)

	return &collisionTestEnv{
		tanksRepo:   tanksRepo,
		bulletsRepo: bulletsRepo,
		mapEntity:   mapEntity,
		bulletUC:    bulletUC,
		tankCommon:  tankCommon,
		tankActions: tankActions,
		collision:   collision,
		entities:    entities,
		lifecycle:   lifecycle,
		specsUC:     specsUC,
	}
}

// tick воспроизводит порядок игрового цикла: движение танков → движение пуль →
// разрешение коллизий
func (env *collisionTestEnv) tick() {
	_ = env.tankCommon.UpdateAllTanks(testDT)
	_ = env.bulletUC.UpdateBullets(testDT)
	env.collision.UpdateCollisions()
}

func (env *collisionTestEnv) newTank(
	role types.TankRole,
	direction types.Direction,
	x, y float64,
	level uint,
) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(role, direction)
	tank := &tankValue
	tank.Position = types.Position{X: x, Y: y}
	tank.PrevPosition = tank.Position
	tank.State = types.TankStateStopped
	tank.SetSpecs(env.specsUC.GetTankSpecs(role == types.TankRoleEnemy, level))
	tank.SetHitPoints(1)
	return tank
}

func brick(x, y float64) *types.BlockEntity {
	return types.NewBlockEntity("brick", x, y, 8, &stubImageProvider{})
}

func water(x, y float64) *types.BlockEntity {
	return types.NewBlockEntity("water", x, y, 8, &stubImageProvider{})
}

// M1: регрессия бага QA — два танка давят друг на друга лоб-в-лоб,
// позиции должны стабилизироваться без осцилляции и перекрытия
func TestTankTankCollision_HeadOnNoJitter(t *testing.T) {
	env := newCollisionTestEnv(nil)

	tankA := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		64,
		100,
		0,
	)
	tankB := env.newTank(types.TankRolePlayer2, types.DirectionLeft, 96, 100, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tankA)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, tankB)

	var prevA, prevB types.Position
	for i := 0; i < 160; i++ {
		// Эмулируем зажатые клавиши: оба игрока продолжают давить
		_ = env.tankActions.Move(tankA)
		_ = env.tankActions.Move(tankB)
		env.tick()

		if env.entities.CheckColliders(tankA, tankB) {
			t.Fatalf(
				"тик %d: танки перекрылись: A=%v B=%v",
				i,
				tankA.Position,
				tankB.Position,
			)
		}

		if i >= 40 {
			if tankA.Position != prevA || tankB.Position != prevB {
				t.Fatalf(
					"тик %d: дрожание — позиции меняются после контакта: A %v→%v, B %v→%v",
					i,
					prevA,
					tankA.Position,
					prevB,
					tankB.Position,
				)
			}
		}
		prevA = tankA.Position
		prevB = tankB.Position
	}

	if tankA.Position.X+float64(tankA.Size.Width) > tankB.Position.X {
		t.Errorf(
			"танки перекрыты по X: A=%v B=%v",
			tankA.Position,
			tankB.Position,
		)
	}
}

// M2: упор в стоящий танк — стоящий не сдвигается, движущийся останавливается
func TestTankTankCollision_MovingIntoStoppedTank(t *testing.T) {
	env := newCollisionTestEnv(nil)

	mover := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		64,
		100,
		0,
	)
	blocker := env.newTank(types.TankRolePlayer2, types.DirectionUp, 96, 100, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, mover)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, blocker)

	blockerStart := blocker.Position

	for i := 0; i < 120; i++ {
		_ = env.tankActions.Move(mover)
		env.tick()

		if blocker.Position != blockerStart {
			t.Fatalf(
				"тик %d: стоящий танк сдвинут с %v на %v",
				i,
				blockerStart,
				blocker.Position,
			)
		}
		if env.entities.CheckColliders(mover, blocker) {
			t.Fatalf("тик %d: перекрытие танков", i)
		}
	}

	if mover.Position.X+float64(mover.Size.Width) > blocker.Position.X {
		t.Errorf("движущийся танк заехал в стоящего: %v", mover.Position)
	}
	if blocker.State != types.TankStateStopped {
		t.Errorf("стоящий танк изменил состояние: %v", blocker.State)
	}
}

// M3: погоня — быстрый догоняющий не наезжает на лидера
func TestTankTankCollision_ChaseNoOverlap(t *testing.T) {
	env := newCollisionTestEnv(nil)

	leader := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		96,
		100,
		0,
	)
	// Враг 1 уровня быстрее игрока (48 против 32)
	chaser := env.newTank(types.TankRoleEnemy, types.DirectionRight, 64, 100, 1)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, leader)
	env.tanksRepo.AddEnemy(chaser)

	prevLeaderX := leader.Position.X
	for i := 0; i < 120; i++ {
		_ = env.tankActions.Move(leader)
		_ = env.tankActions.Move(chaser)
		env.tick()

		if env.entities.CheckColliders(leader, chaser) {
			t.Fatalf(
				"тик %d: догоняющий перекрыл лидера: leader=%v chaser=%v",
				i,
				leader.Position,
				chaser.Position,
			)
		}
		if leader.Position.X < prevLeaderX {
			t.Fatalf("тик %d: лидер откатился назад", i)
		}
		prevLeaderX = leader.Position.X

		if leader.Position.X+float64(leader.Size.Width) >
			float64(208) {
			break // лидер дошёл до края — дальше сценарий не показателен
		}
	}
}

// M4: упор в стену из нескольких блоков по фронту — все блоки разрешаются
// за тик, танк стоит вплотную без дрожания
func TestTankBlockCollision_FlushAgainstWall(t *testing.T) {
	blocks := types.MapBlocks{
		brick(120, 96),
		brick(120, 104),
		brick(120, 112),
	}
	env := newCollisionTestEnv(blocks)

	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 64, 100, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	for i := 0; i < 160; i++ {
		_ = env.tankActions.Move(tank)
		env.tick()

		if i >= 80 && tank.Position.X != 104 {
			t.Fatalf(
				"тик %d: танк не стоит вплотную к стене: X=%v (ожидалось 104)",
				i,
				tank.Position.X,
			)
		}
	}

	if len(env.mapEntity.GetBlocks()) != 3 {
		t.Errorf(
			"танк разрушил блоки: осталось %d",
			len(env.mapEntity.GetBlocks()),
		)
	}
}

// M5: из существующего перекрытия можно выехать, но нельзя углубиться
func TestTankTankCollision_EscapeFromOverlap(t *testing.T) {
	t.Run("выезд разрешён", func(t *testing.T) {
		env := newCollisionTestEnv(nil)

		// Частичное перекрытие: escaper левее, выезжает влево
		escaper := env.newTank(
			types.TankRolePlayer1,
			types.DirectionLeft,
			90,
			100,
			0,
		)
		other := env.newTank(
			types.TankRolePlayer2,
			types.DirectionUp,
			100,
			100,
			0,
		)
		env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, escaper)
		env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, other)

		startX := escaper.Position.X
		for i := 0; i < 60; i++ {
			_ = env.tankActions.Move(escaper)
			env.tick()
		}

		if escaper.Position.X >= startX {
			t.Errorf(
				"танк не смог выехать из перекрытия: X=%v",
				escaper.Position.X,
			)
		}
		if env.entities.CheckColliders(escaper, other) {
			t.Errorf("перекрытие не устранено после выезда")
		}
	})

	t.Run("углубление запрещено", func(t *testing.T) {
		env := newCollisionTestEnv(nil)

		digger := env.newTank(
			types.TankRolePlayer1,
			types.DirectionRight,
			90,
			100,
			0,
		)
		other := env.newTank(
			types.TankRolePlayer2,
			types.DirectionUp,
			100,
			100,
			0,
		)
		env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, digger)
		env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, other)

		for i := 0; i < 60; i++ {
			_ = env.tankActions.Move(digger)
			env.tick()

			if digger.Position.X > 90 {
				t.Fatalf(
					"тик %d: танк углубился в перекрытие: X=%v",
					i,
					digger.Position.X,
				)
			}
		}
	})
}

// B1: два попадания за один тик — удаляются именно попавшие пули
func TestBulletTankCollision_TwoHitsSameTick(t *testing.T) {
	env := newCollisionTestEnv(nil)

	shooter1 := env.newTank(types.TankRolePlayer1, types.DirectionUp, 0, 192, 0)
	shooter2 := env.newTank(
		types.TankRolePlayer2,
		types.DirectionUp,
		32,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter1)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, shooter2)

	// Тяжёлые враги (2 hp), чтобы проверить и удаление пуль, и урон
	enemy1 := env.newTank(types.TankRoleEnemy, types.DirectionDown, 64, 64, 0)
	enemy1.SetHitPoints(2)
	enemy2 := env.newTank(types.TankRoleEnemy, types.DirectionDown, 128, 64, 0)
	enemy2.SetHitPoints(2)
	env.tanksRepo.AddEnemy(enemy1)
	env.tanksRepo.AddEnemy(enemy2)

	specs := env.specsUC.GetTankSpecs(false, 0)
	bullet1 := types.NewBulletEntity(
		types.Position{X: 70, Y: 70},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionUp,
		specs,
		shooter1,
	)
	bullet2 := types.NewBulletEntity(
		types.Position{X: 134, Y: 70},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionUp,
		specs,
		shooter2,
	)
	if err := env.bulletsRepo.AddBullet(bullet1); err != nil {
		t.Fatalf("не удалось добавить пулю 1: %v", err)
	}
	if err := env.bulletsRepo.AddBullet(bullet2); err != nil {
		t.Fatalf("не удалось добавить пулю 2: %v", err)
	}

	env.collision.UpdateCollisions()

	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("ожидалось 0 пуль после двух попаданий, осталось %d", got)
	}
	if enemy1.GetHitPoints() != 1 {
		t.Errorf("враг 1 не получил урон: hp=%d", enemy1.GetHitPoints())
	}
	if enemy2.GetHitPoints() != 1 {
		t.Errorf("враг 2 не получил урон: hp=%d", enemy2.GetHitPoints())
	}
}

// Вода не останавливает пули: пуля над водой выживает, блок цел
func TestBulletWallCollision_BulletPassesOverWater(t *testing.T) {
	blocks := types.MapBlocks{water(120, 96)}
	env := newCollisionTestEnv(blocks)

	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)

	specs := env.specsUC.GetTankSpecs(false, 0)
	bullet := types.NewBulletEntity(
		types.Position{X: 121, Y: 97},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionRight,
		specs,
		shooter,
	)
	if err := env.bulletsRepo.AddBullet(bullet); err != nil {
		t.Fatalf("не удалось добавить пулю: %v", err)
	}

	env.collision.UpdateCollisions()

	if got := len(env.bulletUC.GetBullets()); got != 1 {
		t.Errorf("пуля погибла над водой: осталось %d пуль", got)
	}
	if got := len(env.mapEntity.GetBlocks()); got != 1 {
		t.Errorf("вода удалена: осталось %d блоков", got)
	}
}

// Вода блокирует танки как обычная стена
func TestTankWallCollision_WaterBlocksTank(t *testing.T) {
	blocks := types.MapBlocks{water(120, 96), water(120, 104)}
	env := newCollisionTestEnv(blocks)

	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 96, 96, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	for i := 0; i < 40; i++ {
		_ = env.tankActions.Move(tank)
		env.tick()
	}

	if tank.Position.X != 104 {
		t.Errorf("танк не упёрся в воду: X=%v, ожидалось 104", tank.Position.X)
	}
}

// B2/B3: два выстрела в один кирпич за тик — кирпич удаляется один раз
// (без звукового сервиса), вторая пуля летит дальше, а не бьёт фантом
func TestBulletWallCollision_NoPhantomHitOnDestroyedBrick(t *testing.T) {
	blocks := types.MapBlocks{brick(120, 96)}
	env := newCollisionTestEnv(blocks)

	shooter1 := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	shooter2 := env.newTank(
		types.TankRolePlayer2,
		types.DirectionRight,
		32,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter1)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, shooter2)

	specs := env.specsUC.GetTankSpecs(false, 0)
	// Обе пули пересекают один и тот же кирпич, но не друг друга
	hitting := types.NewBulletEntity(
		types.Position{X: 121, Y: 97},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionRight,
		specs,
		shooter1,
	)
	trailing := types.NewBulletEntity(
		types.Position{X: 118, Y: 101.5},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionRight,
		specs,
		shooter2,
	)
	if err := env.bulletsRepo.AddBullet(hitting); err != nil {
		t.Fatalf("не удалось добавить пулю: %v", err)
	}
	if err := env.bulletsRepo.AddBullet(trailing); err != nil {
		t.Fatalf("не удалось добавить пулю: %v", err)
	}

	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf(
			"кирпич не удалён (soundUseCases == nil): осталось %d блоков",
			got,
		)
	}

	remaining := env.bulletUC.GetBullets()
	if len(remaining) != 1 {
		t.Fatalf(
			"ожидалась 1 выжившая пуля (вторая не должна бить фантомный кирпич), осталось %d",
			len(remaining),
		)
	}
	if remaining[0] != trailing {
		t.Errorf("выжила не та пуля")
	}
}
