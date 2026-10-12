package use_cases_test

import (
	"image/color"
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/services/collision_services"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
	"github.com/shpaker/tnk9x/internal/use_cases/tank_use_cases"
)

const testDT = 1.0 / 60.0

type stubImageProvider struct{}

func (s *stubImageProvider) GetImageID() (string, error) { return "stub", nil }

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

func (s *stubRenderUseCases) IsTankBlinking(tank *types.TankEntity) bool {
	return false
}

func (s *stubRenderUseCases) TankTint(
	tank *types.TankEntity,
) (color.NRGBA, bool) {
	return color.NRGBA{}, false
}

type stubHQUseCases struct {
	hq         *types.HQEntity
	explosions int
}

func (s *stubHQUseCases) GetHQ() *types.HQEntity { return s.hq }
func (s *stubHQUseCases) Explode(hq *types.HQEntity) error {
	s.explosions++
	return nil
}

func (s *stubHQUseCases) IsExplosionFinished(hq *types.HQEntity) {
}
func (s *stubHQUseCases) IsDestroyed() bool { return false }

type stubTankLifecycle struct {
	exploded []*types.TankEntity
}

func (s *stubTankLifecycle) SpawnEnemy(
	spawnerIndex int,
	level uint,
) (*types.TankEntity, error) {
	return nil, nil
}

func (s *stubTankLifecycle) SpawnPlayer1(
	uint,
) (*types.TankEntity, error) {
	return nil, nil
}

func (s *stubTankLifecycle) SpawnPlayer2(
	uint,
) (*types.TankEntity, error) {
	return nil, nil
}

func (s *stubTankLifecycle) CompleteSpawn(*types.TankEntity) {}

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
func (s *stubMapUseCases) RestoreBlock(
	*types.BlockEntity, []types.IEntityCollider,
) bool {
	return false
}

func (s *stubMapUseCases) IsBlockIntact(*types.BlockEntity) bool { return true }

func (s *stubMapUseCases) RemoveBlock(block *types.BlockEntity) error {
	return s.mapEntity.RemoveBlock(block)
}

func (s *stubMapUseCases) GetSizePx() types.Size { return s.mapEntity.GetSizePx() }

func (s *stubMapUseCases) GetBorders() []*types.BorderEntity {
	return s.mapEntity.GetBorders()
}

func (s *stubMapUseCases) GetRandomBonusSpawnPosition() types.Position {
	return types.Position{}
}

func (s *stubMapUseCases) IsIceAt(_ types.Position) bool { return false }

func (s *stubMapUseCases) IsWaterUnder(
	position types.Position,
	size types.Size,
) bool {
	return use_cases.NewMapUseCases(s.mapEntity).IsWaterUnder(position, size)
}

type recordingBonusUseCases struct {
	takenBy []*types.TankEntity
}

func (r *recordingBonusUseCases) Apply(
	_ *types.BonusEntity,
	tank *types.TankEntity,
) {
	r.takenBy = append(r.takenBy, tank)
}

func (r *recordingBonusUseCases) UpdateEffects() {}

func (r *recordingBonusUseCases) SpawnRandomBonusEntity(
	types.Position,
) *types.BonusEntity {
	return nil
}

func (r *recordingBonusUseCases) VisibleBonuses() []*types.BonusEntity {
	return nil
}

type collisionTestEnv struct {
	bonuses     *recordingBonusUseCases
	bonusesRepo *game.BonusesRepository
	tanksRepo   *game.TanksRepository
	bulletsRepo *game.BulletsRepository
	mapEntity   *types.MapEntity
	bulletUC    *use_cases.BulletUseCases
	tankCommon  *tank_use_cases.TankCommonUseCases
	tankActions *tank_use_cases.TankActionsUseCases
	collision   *use_cases.CollisionUseCases
	lifecycle   *stubTankLifecycle
	specsUC     *use_cases.SpecsUseCases
	soundUC     *use_cases.SoundUseCases
	effects     *testutil.FakeVisualEffectsUseCases
	hqUseCases  *stubHQUseCases
}

func newCollisionTestEnv(blocks types.MapBlocks) *collisionTestEnv {
	return newCollisionTestEnvWith(blocks, false)
}

func newCollisionTestEnvWith(
	blocks types.MapBlocks,
	enemyBonusPickup bool,
) *collisionTestEnv {
	tanksRepo := game.NewTanksRepository()
	bulletsRepo := game.NewBulletsRepository()
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		blocks,
		nil,
	)
	mapUC := &stubMapUseCases{mapEntity: mapEntity}

	specsUC := use_cases.NewSpecsUseCases()

	render := &stubRenderUseCases{}
	soundUC := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())
	effects := &testutil.FakeVisualEffectsUseCases{}

	bulletUC := use_cases.NewBulletUseCases(bulletsRepo, nil, 16, false)
	tankCommon := tank_use_cases.NewTankCommonUseCases(
		render,
		tanksRepo,
		specsUC,
		mapUC,
		session_entities.NewStageSessionEntity(),
	)
	tankActions := tank_use_cases.NewTankActionsUseCases(
		bulletUC,
		tankCommon,
		render,
		soundUC,
		effects,
	)
	lifecycle := &stubTankLifecycle{}

	bonuses := &recordingBonusUseCases{}
	bonusesRepo := game.NewBonusesRepository()
	hqUseCases := &stubHQUseCases{}
	collision := use_cases.NewCollisionUseCases(
		bulletUC,
		tankActions,
		mapUC,
		tankCommon,
		lifecycle,
		collision_services.NewTankCollisionService(enemyBonusPickup),
		collision_services.NewBulletCollisionService(),
		collision_services.NewSpawnCollisionService(),
		hqUseCases,
		bonuses,
		bonusesRepo,
		game.NewCollisionWorldRepository(),
		soundUC,
		effects,
	)

	return &collisionTestEnv{
		bonuses:     bonuses,
		bonusesRepo: bonusesRepo,
		tanksRepo:   tanksRepo,
		bulletsRepo: bulletsRepo,
		mapEntity:   mapEntity,
		bulletUC:    bulletUC,
		tankCommon:  tankCommon,
		tankActions: tankActions,
		collision:   collision,
		lifecycle:   lifecycle,
		specsUC:     specsUC,
		soundUC:     soundUC,
		effects:     effects,
		hqUseCases:  hqUseCases,
	}
}

// Mirrors the game loop order: tank movement, bullet movement,
// then collision resolution.
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
	tank.State = types.TankStateActive
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

func steel(x, y float64) *types.BlockEntity {
	return types.NewBlockEntity("steel", x, y, 8, &stubImageProvider{})
}

func brickRemnant(
	dataX, dataY, x, y float64,
	w, h int,
) *types.BlockEntity {
	block := types.NewBlockEntity(
		"brick",
		dataX,
		dataY,
		8,
		&stubImageProvider{},
	)
	block.Position = types.Position{X: x, Y: y}
	block.Size = types.Size{Width: w, Height: h}
	return block
}

func (env *collisionTestEnv) addBullet(
	t *testing.T,
	x, y float64,
	direction types.Direction,
	level uint,
	owner *types.TankEntity,
) *types.BulletEntity {
	t.Helper()
	bullet := types.NewBulletEntity(
		types.Position{X: x, Y: y},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		direction,
		env.specsUC.GetTankSpecs(false, level),
		owner,
	)
	if err := env.bulletsRepo.AddBullet(bullet); err != nil {
		t.Fatalf("failed to add bullet: %v", err)
	}
	return bullet
}

func countSounds(events []types.SoundEntity, id types.SoundID) int {
	count := 0
	for _, event := range events {
		if event.SoundID == id {
			count++
		}
	}
	return count
}

func assertBlockRect(
	t *testing.T,
	block *types.BlockEntity,
	x, y float64,
	w, h int,
) {
	t.Helper()
	if block.Position.X != x || block.Position.Y != y ||
		block.Size.Width != w || block.Size.Height != h {
		t.Errorf(
			"block remnant (%v, %dx%d), want ({%v %v}, %dx%d)",
			block.Position,
			block.Size.Width,
			block.Size.Height,
			x,
			y,
			w,
			h,
		)
	}
}

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
		_ = env.tankActions.Move(tankA)
		_ = env.tankActions.Move(tankB)
		env.tick()

		if types.Overlaps(tankA, tankB) {
			t.Fatalf(
				"tick %d: tanks overlap: A=%v B=%v",
				i,
				tankA.Position,
				tankB.Position,
			)
		}

		if i >= 40 {
			if tankA.Position != prevA || tankB.Position != prevB {
				t.Fatalf(
					"tick %d: jitter, positions change after contact: A %v→%v, B %v→%v",
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
			"tanks overlap on X: A=%v B=%v",
			tankA.Position,
			tankB.Position,
		)
	}
}

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
				"tick %d: stationary tank moved from %v to %v",
				i,
				blockerStart,
				blocker.Position,
			)
		}
		if types.Overlaps(mover, blocker) {
			t.Fatalf("tick %d: tanks overlap", i)
		}
	}

	if mover.Position.X+float64(mover.Size.Width) > blocker.Position.X {
		t.Errorf(
			"moving tank drove into the stationary one: %v",
			mover.Position,
		)
	}
	if blocker.State != types.TankStateActive {
		t.Errorf("stationary tank changed state: %v", blocker.State)
	}
}

func TestTankTankCollision_ChaseNoOverlap(t *testing.T) {
	env := newCollisionTestEnv(nil)

	leader := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		96,
		100,
		0,
	)
	// A level-1 enemy is faster than the player (48 vs 32).
	chaser := env.newTank(types.TankRoleEnemy, types.DirectionRight, 64, 100, 1)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, leader)
	env.tanksRepo.AddEnemy(chaser)

	prevLeaderX := leader.Position.X
	for i := 0; i < 120; i++ {
		_ = env.tankActions.Move(leader)
		_ = env.tankActions.Move(chaser)
		env.tick()

		if types.Overlaps(leader, chaser) {
			t.Fatalf(
				"tick %d: chaser overlaps leader: leader=%v chaser=%v",
				i,
				leader.Position,
				chaser.Position,
			)
		}
		if leader.Position.X < prevLeaderX {
			t.Fatalf("tick %d: leader rolled back", i)
		}
		prevLeaderX = leader.Position.X

		if leader.Position.X+float64(leader.Size.Width) >
			float64(208) {
			break // the leader reached the edge; the rest is not representative
		}
	}
}

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
				"tick %d: tank is not flush with the wall: X=%v (want 104)",
				i,
				tank.Position.X,
			)
		}
	}

	if len(env.mapEntity.GetBlocks()) != 3 {
		t.Errorf(
			"tank destroyed blocks: %d left",
			len(env.mapEntity.GetBlocks()),
		)
	}
}

func TestTankHQCollision_IntactHQBlocksTank(t *testing.T) {
	env := newCollisionTestEnv(nil)
	env.hqUseCases.hq = &types.HQEntity{
		Position: types.Position{X: 120, Y: 100},
	}

	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 64, 100, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	for i := 0; i < 160; i++ {
		_ = env.tankActions.Move(tank)
		env.tick()
	}

	if tank.Position.X != 104 {
		t.Errorf("tank is not flush with the HQ: X=%v (want 104)",
			tank.Position.X)
	}
}

func TestTankHQCollision_DestroyedHQPassable(t *testing.T) {
	env := newCollisionTestEnv(nil)
	env.hqUseCases.hq = &types.HQEntity{
		Position: types.Position{X: 120, Y: 100},
		State:    types.HQStateDestroyed,
	}

	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 64, 100, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	for i := 0; i < 160; i++ {
		_ = env.tankActions.Move(tank)
		env.tick()
	}

	if tank.Position.X <= 120 {
		t.Errorf("tank did not pass the destroyed HQ: X=%v", tank.Position.X)
	}
}

func TestTankTankCollision_EscapeFromOverlap(t *testing.T) {
	t.Run("escape allowed", func(t *testing.T) {
		env := newCollisionTestEnv(nil)

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
				"tank could not escape the overlap: X=%v",
				escaper.Position.X,
			)
		}
		if types.Overlaps(escaper, other) {
			t.Errorf("overlap remains after escaping")
		}
	})

	t.Run("going deeper forbidden", func(t *testing.T) {
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
					"tick %d: tank went deeper into the overlap: X=%v",
					i,
					digger.Position.X,
				)
			}
		}
	})
}

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
		t.Fatalf("failed to add bullet 1: %v", err)
	}
	if err := env.bulletsRepo.AddBullet(bullet2); err != nil {
		t.Fatalf("failed to add bullet 2: %v", err)
	}

	env.collision.UpdateCollisions()

	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("want 0 bullets after two hits, got %d", got)
	}
	if enemy1.GetHitPoints() != 1 {
		t.Errorf("enemy 1 took no damage: hp=%d", enemy1.GetHitPoints())
	}
	if enemy2.GetHitPoints() != 1 {
		t.Errorf("enemy 2 took no damage: hp=%d", enemy2.GetHitPoints())
	}
}

func TestBulletTankCollision_ShieldAbsorbsHit(t *testing.T) {
	env := newCollisionTestEnv(nil)

	player := env.newTank(types.TankRolePlayer1, types.DirectionUp, 64, 64, 1)
	player.ActivateShield(600)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, player)

	enemy := env.newTank(types.TankRoleEnemy, types.DirectionDown, 64, 160, 0)
	env.tanksRepo.AddEnemy(enemy)

	bullet := types.NewBulletEntity(
		types.Position{X: 70, Y: 70},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionUp,
		env.specsUC.GetTankSpecs(true, 0),
		enemy,
	)
	if err := env.bulletsRepo.AddBullet(bullet); err != nil {
		t.Fatalf("failed to add bullet: %v", err)
	}

	env.collision.UpdateCollisions()

	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet not absorbed by shield: %d left", got)
	}
	if player.GetSpecs().GetLevel() != 1 {
		t.Errorf(
			"player level changed under shield: %d",
			player.GetSpecs().GetLevel(),
		)
	}
	if player.IsDestroyed() {
		t.Error("player destroyed under shield")
	}
	if kinds := env.effects.Kinds(); len(kinds) != 1 ||
		kinds[0] != types.VisualEventShieldHit {
		t.Errorf("effect events %v, want shield hit", kinds)
	}
}

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
		t.Fatalf("failed to add bullet: %v", err)
	}

	env.collision.UpdateCollisions()

	if got := len(env.bulletUC.GetBullets()); got != 1 {
		t.Errorf("bullet died over water: %d left", got)
	}
	if got := len(env.mapEntity.GetBlocks()); got != 1 {
		t.Errorf("water removed: %d blocks left", got)
	}
}

func TestBulletWallCollision_BrickShieldsHQFromSameTickHit(t *testing.T) {
	env := newCollisionTestEnv(types.MapBlocks{brick(112, 96)})
	env.hqUseCases.hq = &types.HQEntity{
		Position: types.Position{X: 120, Y: 100},
	}

	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)
	env.addBullet(t, 118, 102, types.DirectionRight, 0, shooter)

	env.collision.UpdateCollisions()

	if env.hqUseCases.explosions != 0 {
		t.Errorf("HQ blown up by a bullet the brick should have absorbed")
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet not extinguished: %d left", got)
	}
}

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
		t.Errorf("tank did not stop at water: X=%v, want 104", tank.Position.X)
	}
}

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
		t.Fatalf("failed to add bullet: %v", err)
	}
	if err := env.bulletsRepo.AddBullet(trailing); err != nil {
		t.Fatalf("failed to add bullet: %v", err)
	}

	env.collision.UpdateCollisions()

	blocksLeft := env.mapEntity.GetBlocks()
	if len(blocksLeft) != 1 {
		t.Fatalf(
			"want 1 brick remnant, got %d blocks",
			len(blocksLeft),
		)
	}
	assertBlockRect(t, blocksLeft[0], 124, 96, 4, 8)

	remaining := env.bulletUC.GetBullets()
	if len(remaining) != 1 {
		t.Fatalf(
			"want 1 surviving bullet (the second must not hit a phantom brick), got %d",
			len(remaining),
		)
	}
	if remaining[0] != trailing {
		t.Errorf("wrong bullet survived")
	}
}

func TestBulletWallCollision_BrickShavedByDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction types.Direction
		bulletX   float64
		bulletY   float64
		wantX     float64
		wantY     float64
		wantW     int
		wantH     int
		debrisX   float64
		debrisY   float64
	}{
		{"right", types.DirectionRight, 117, 98, 124, 96, 4, 8, 120, 96},
		{"left", types.DirectionLeft, 127, 98, 120, 96, 4, 8, 124, 96},
		{"down", types.DirectionDown, 122, 93, 120, 100, 8, 4, 120, 96},
		{"up", types.DirectionUp, 122, 103, 120, 96, 8, 4, 120, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newCollisionTestEnv(types.MapBlocks{brick(120, 96)})
			shooter := env.newTank(
				types.TankRolePlayer1,
				tt.direction,
				0,
				192,
				0,
			)
			env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)
			env.addBullet(t, tt.bulletX, tt.bulletY, tt.direction, 0, shooter)

			env.collision.UpdateCollisions()

			blocks := env.mapEntity.GetBlocks()
			if len(blocks) != 1 {
				t.Fatalf("want 1 remnant, got %d blocks", len(blocks))
			}
			assertBlockRect(
				t,
				blocks[0],
				tt.wantX,
				tt.wantY,
				tt.wantW,
				tt.wantH,
			)
			if got := len(env.bulletUC.GetBullets()); got != 0 {
				t.Errorf("bullet not removed: %d", got)
			}
			events := env.soundUC.GetEvents()
			if got := countSounds(events, types.SoundIDBrick); got != 1 {
				t.Errorf("brick sounds %d, want 1", got)
			}
			debris := env.effects.Events[0]
			want := types.Position{X: tt.debrisX, Y: tt.debrisY}
			if debris.Kind != types.VisualEventBlockDebris ||
				debris.Position != want ||
				!isSlab(debris.Size, tt.wantW, tt.wantH) {
				t.Errorf("debris %v at %v of size %v", debris.Kind,
					debris.Position, debris.Size)
			}
		})
	}
}

func TestBulletWallCollision_SecondHitSameDirectionDestroys(t *testing.T) {
	env := newCollisionTestEnv(types.MapBlocks{brick(120, 96)})
	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)

	env.addBullet(t, 117, 98, types.DirectionRight, 0, shooter)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 1 {
		t.Fatalf("want a remnant after the first shot, got %d blocks", got)
	}

	env.addBullet(t, 121, 98, types.DirectionRight, 0, shooter)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf("second shot did not finish the tile: %d blocks", got)
	}
}

func TestBulletWallCollision_SideHitOnHalfLeavesQuarter(t *testing.T) {
	env := newCollisionTestEnv(
		types.MapBlocks{brickRemnant(120, 96, 120, 100, 8, 4)},
	)
	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)

	env.addBullet(t, 117, 101, types.DirectionRight, 0, shooter)
	env.collision.UpdateCollisions()

	blocks := env.mapEntity.GetBlocks()
	if len(blocks) != 1 {
		t.Fatalf("want a quarter, got %d blocks", len(blocks))
	}
	assertBlockRect(t, blocks[0], 124, 100, 4, 4)
}

// A remnant thin along the bullet path is destroyed entirely.
func TestBulletWallCollision_HitFromOpenSideDestroysThinRemnant(
	t *testing.T,
) {
	env := newCollisionTestEnv(
		types.MapBlocks{brickRemnant(120, 96, 124, 96, 4, 8)},
	)
	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)

	env.addBullet(t, 121, 98, types.DirectionRight, 0, shooter)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf("thin remnant not destroyed: %d blocks", got)
	}
}

func TestBulletWallCollision_ReinforcedDestroysWholeBrick(t *testing.T) {
	blocks := types.MapBlocks{
		brick(120, 96),
		brickRemnant(120, 120, 124, 120, 4, 8),
	}
	env := newCollisionTestEnv(blocks)
	shooter1 := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		2,
	)
	shooter2 := env.newTank(
		types.TankRolePlayer2,
		types.DirectionRight,
		32,
		192,
		2,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter1)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, shooter2)

	env.addBullet(t, 117, 98, types.DirectionRight, 2, shooter1)
	env.addBullet(t, 121, 122, types.DirectionRight, 2, shooter2)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf(
			"reinforced bullets did not destroy the whole brick: %d blocks",
			got,
		)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullets not removed: %d", got)
	}
}

func TestBulletWallCollision_SeamStraddleShavesBothTiles(t *testing.T) {
	top := brick(120, 96)
	bottom := brick(120, 104)
	env := newCollisionTestEnv(types.MapBlocks{top, bottom})
	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)

	env.addBullet(t, 117, 102, types.DirectionRight, 0, shooter)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 2 {
		t.Fatalf("want 2 remnants, got %d blocks", got)
	}
	assertBlockRect(t, top, 124, 96, 4, 8)
	assertBlockRect(t, bottom, 124, 104, 4, 8)
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet not removed: %d", got)
	}
	events := env.soundUC.GetEvents()
	if got := countSounds(events, types.SoundIDBrick); got != 1 {
		t.Errorf("brick sounds %d, want 1", got)
	}
}

func TestBulletWallCollision_BrickAndSteelSeam(t *testing.T) {
	brickBlock := brick(120, 96)
	steelBlock := steel(120, 104)
	env := newCollisionTestEnv(types.MapBlocks{brickBlock, steelBlock})
	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)

	env.addBullet(t, 117, 102, types.DirectionRight, 0, shooter)
	env.collision.UpdateCollisions()

	assertBlockRect(t, brickBlock, 124, 96, 4, 8)
	assertBlockRect(t, steelBlock, 120, 104, 8, 8)
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet not removed: %d", got)
	}
	events := env.soundUC.GetEvents()
	if got := countSounds(events, types.SoundIDBrick); got != 1 {
		t.Errorf("brick sounds %d, want 1", got)
	}
	if got := countSounds(events, types.SoundIDSteel); got != 1 {
		t.Errorf("steel sounds %d, want 1", got)
	}
	kinds := env.effects.Kinds()
	if len(kinds) != 3 || kinds[0] != types.VisualEventBlockDebris ||
		kinds[1] != types.VisualEventBrickHit ||
		kinds[2] != types.VisualEventSteelHit {
		t.Fatalf(
			"effect events %v, want debris, brick and steel",
			kinds,
		)
	}
	hit := env.effects.Events[1]
	if hit.Position != (types.Position{X: 119, Y: 104}) ||
		hit.Direction != types.DirectionRight {
		t.Errorf("effect at %v along %v", hit.Position, hit.Direction)
	}
}

func TestBulletWallCollision_SteelBouncesNormalBullet(t *testing.T) {
	env := newCollisionTestEnv(types.MapBlocks{steel(120, 96)})
	normal := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	reinforced := env.newTank(
		types.TankRolePlayer2,
		types.DirectionRight,
		32,
		192,
		2,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, normal)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer2, reinforced)

	env.addBullet(t, 117, 98, types.DirectionRight, 0, normal)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 1 {
		t.Fatalf("normal bullet damaged steel: %d blocks", got)
	}
	events := env.soundUC.GetEvents()
	if got := countSounds(events, types.SoundIDSteel); got != 1 {
		t.Errorf("steel sounds %d, want 1", got)
	}

	env.addBullet(t, 117, 98, types.DirectionRight, 2, reinforced)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf("reinforced bullet did not destroy steel: %d blocks", got)
	}
}

func TestBulletWallCollision_TwoBulletsSameTickFinishTile(t *testing.T) {
	env := newCollisionTestEnv(types.MapBlocks{brick(120, 96)})
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

	// Bullets touch the brick but not each other (touching is not a collision).
	env.addBullet(t, 121, 97, types.DirectionRight, 0, shooter1)
	env.addBullet(t, 121, 101, types.DirectionRight, 0, shooter2)
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf("tile not finished by the second bullet: %d blocks", got)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullets not removed: %d", got)
	}
}

func isSlab(size types.Size, remainW, remainH int) bool {
	if remainW < 8 {
		return size == types.Size{Width: 8 - remainW, Height: remainH}
	}
	return size == types.Size{Width: remainW, Height: 8 - remainH}
}

func TestBulletBulletCollision_Effect(t *testing.T) {
	env := newCollisionTestEnv(nil)
	player := env.newTank(types.TankRolePlayer1, types.DirectionUp, 0, 96, 0)
	enemy := env.newTank(types.TankRoleEnemy, types.DirectionDown, 96, 0, 0)

	env.addBullet(t, 100, 100, types.DirectionUp, 0, player)
	env.addBullet(t, 100, 98, types.DirectionDown, 0, enemy)
	env.collision.UpdateCollisions()

	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullets left %d, want 0", got)
	}
	kinds := env.effects.Kinds()
	if len(kinds) != 1 || kinds[0] != types.VisualEventBulletClash {
		t.Fatalf("effects %v, want one bullet clash", kinds)
	}
	if got := env.effects.Events[0].Position; got != (types.Position{X: 102, Y: 101}) {
		t.Errorf("clash at %v, want the midpoint (102, 101)", got)
	}
}

func TestTankWallCollision_BoatSailsOverWater(t *testing.T) {
	blocks := types.MapBlocks{water(120, 96), water(120, 104)}
	env := newCollisionTestEnv(blocks)

	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 96, 96, 0)
	tank.SetBoat()
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	for i := 0; i < 40; i++ {
		_ = env.tankActions.Move(tank)
		env.tick()
	}

	if tank.Position.X <= 104 {
		t.Errorf("tank with boat did not enter water: X=%v", tank.Position.X)
	}
}

func TestBulletTankCollision_BoatAbsorbsHit(t *testing.T) {
	env := newCollisionTestEnv(nil)

	player := env.newTank(types.TankRolePlayer1, types.DirectionUp, 64, 64, 1)
	player.SetBoat()
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, player)

	enemy := env.newTank(types.TankRoleEnemy, types.DirectionDown, 64, 160, 0)
	env.tanksRepo.AddEnemy(enemy)

	bullet := types.NewBulletEntity(
		types.Position{X: 70, Y: 70},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		&stubImageProvider{},
		types.DirectionUp,
		env.specsUC.GetTankSpecs(true, 0),
		enemy,
	)
	if err := env.bulletsRepo.AddBullet(bullet); err != nil {
		t.Fatalf("failed to add bullet: %v", err)
	}

	env.collision.UpdateCollisions()

	if player.HasBoat() {
		t.Error("boat not knocked out by the hit")
	}
	if player.GetSpecs().GetLevel() != 1 {
		t.Errorf("player level changed: %d", player.GetSpecs().GetLevel())
	}
}

func TestTankWallCollision_SunkBoatReachesShore(t *testing.T) {
	blocks := types.MapBlocks{
		water(104, 96), water(104, 104), water(112, 96), water(112, 104),
	}
	env := newCollisionTestEnv(blocks)

	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 104, 96, 0)
	tank.SetBoat()
	tank.LoseBoat()
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	for i := 0; i < 60; i++ {
		_ = env.tankActions.Move(tank)
		env.tick()
	}
	if tank.Position.X < 120 {
		t.Fatalf("tank did not reach the shore: X=%v", tank.Position.X)
	}
	if tank.CanSail() {
		t.Fatal("tank must not sail on the shore")
	}

	_ = env.tankActions.Rotate(tank, types.DirectionLeft)
	for i := 0; i < 60; i++ {
		_ = env.tankActions.Rotate(tank, types.DirectionLeft)
		_ = env.tankActions.Move(tank)
		env.tick()
	}
	if tank.Position.X < 120 {
		t.Errorf("boatless tank entered water: X=%v", tank.Position.X)
	}
}

func TestTankBonusCollision_EnemyPickupFlag(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		env := newCollisionTestEnvWith(nil, enabled)
		enemy := env.newTank(
			types.TankRoleEnemy,
			types.DirectionDown,
			64,
			64,
			0,
		)
		env.tanksRepo.AddEnemy(enemy)
		env.bonusesRepo.AddBonus(types.NewBonusEntity(
			types.BonusTypeStar,
			types.Position{X: 64, Y: 64},
			types.Size{Width: 16, Height: 16},
			nil,
		))

		env.collision.UpdateCollisions()

		if picked := len(env.bonuses.takenBy) == 1; picked != enabled {
			t.Errorf("pickup enabled=%v, enemy picked=%v", enabled, picked)
		}
	}
}

func TestBulletWallCollision_LongStepHitsThinRemnant(t *testing.T) {
	remnant := brickRemnant(120, 96, 124, 96, 4, 8)
	env := newCollisionTestEnv(types.MapBlocks{remnant})
	shooter := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		0,
		192,
		0,
	)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, shooter)
	bullet := env.addBullet(t, 100, 98, types.DirectionRight, 0, shooter)

	_ = env.bulletUC.UpdateBullets(0.5)
	if bullet.Position.X < 140 {
		t.Fatalf(
			"bullet flew only to X=%v, step too small for the check",
			bullet.Position.X,
		)
	}
	env.collision.UpdateCollisions()

	if got := len(env.mapEntity.GetBlocks()); got != 0 {
		t.Errorf("brick remnant intact: %d blocks", got)
	}
	if got := len(env.bulletUC.GetBullets()); got != 0 {
		t.Errorf("bullet flew through the brick: %d bullets", got)
	}
}

func TestTankBlockCollision_LongStepStopsFlush(t *testing.T) {
	env := newCollisionTestEnv(types.MapBlocks{brick(120, 96), brick(120, 104)})
	tank := env.newTank(types.TankRolePlayer1, types.DirectionRight, 64, 96, 0)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	tank.Position.X = 180
	env.collision.UpdateCollisions()

	if tank.Position.X != 104 {
		t.Errorf(
			"tank did not stop flush with the wall: X=%v, want 104",
			tank.Position.X,
		)
	}
}

func TestTankBorderCollision_HaltsDocking(t *testing.T) {
	env := newCollisionTestEnv(nil)
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		190,
		100,
		0,
	)
	testutil.DockingTank(tank)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	tank.Position.X = 193
	env.collision.UpdateCollisions()

	if !tank.IsStopped() || tank.Position.X != 192 {
		t.Fatalf(
			"stopped=%v X=%v, want stop at 192",
			tank.IsStopped(),
			tank.Position.X,
		)
	}
}

func TestTankBorderCollision_KeepsDriving(t *testing.T) {
	env := newCollisionTestEnv(nil)
	tank := env.newTank(
		types.TankRolePlayer1,
		types.DirectionRight,
		190,
		100,
		0,
	)
	testutil.MovingTank(tank)
	env.tanksRepo.SetPlayer(types.PlayerTankNumPlayer1, tank)

	tank.Position.X = 193
	env.collision.UpdateCollisions()

	if !tank.IsDriving() || tank.Position.X != 192 {
		t.Fatalf(
			"driving=%v X=%v, want driving at the edge at 192",
			tank.IsDriving(),
			tank.Position.X,
		)
	}
}
