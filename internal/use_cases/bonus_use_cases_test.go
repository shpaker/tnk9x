package use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// recordingTankCommon записывает вызовы LevelUp и отдаёт канонный список танков
type recordingTankCommon struct {
	tanks     []*types.TankEntity
	leveledUp []*types.TankEntity
}

func (s *recordingTankCommon) Update(
	tank *types.TankEntity,
	dt float64,
) error {
	return nil
}
func (s *recordingTankCommon) UpdateAllTanks(dt float64) error { return nil }

func (s *recordingTankCommon) GetAllTanks() []*types.TankEntity {
	return s.tanks
}

func (s *recordingTankCommon) GetAllPlayerTanks() []*types.TankEntity {
	return nil
}
func (s *recordingTankCommon) IsAnyPlayerTankMoving() bool { return false }

func (s *recordingTankCommon) LevelUp(tank *types.TankEntity) {
	s.leveledUp = append(s.leveledUp, tank)
}
func (s *recordingTankCommon) LevelDown(tank *types.TankEntity) {}

func (s *recordingTankCommon) GetTankAnimationName(
	tank *types.TankEntity,
) string {
	return ""
}

// recordingLifecycle записывает вызовы Explode
type recordingLifecycle struct {
	exploded []*types.TankEntity
}

func (s *recordingLifecycle) OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error) {
	return [3]*types.TankEntity{}, nil
}

func (s *recordingLifecycle) SpawnEnemyWithLevel(
	index *int,
	ignoreRespawnDelay bool,
	remainingEnemies uint,
) (*types.TankEntity, error) {
	return nil, nil
}

func (s *recordingLifecycle) SpawnPlayer1() (*types.TankEntity, error) {
	return nil, nil
}

func (s *recordingLifecycle) SpawnPlayer2() (*types.TankEntity, error) {
	return nil, nil
}

func (s *recordingLifecycle) GetPlayerTank(
	num types.PlayerTankNum,
) *types.TankEntity {
	return nil
}

func (s *recordingLifecycle) SetPlayerTank(
	num types.PlayerTankNum,
	tank *types.TankEntity,
) {
}

func (s *recordingLifecycle) Explode(tank *types.TankEntity) error {
	s.exploded = append(s.exploded, tank)
	tank.State = types.TankStateExploding
	return nil
}

func (s *recordingLifecycle) UpdateAllTanksLifecycle() error { return nil }

// stubConfigProvider отдаёт только размер базового тайла
type stubConfigProvider struct {
	baseSizePx uint
}

func (s *stubConfigProvider) GetEnemySpawners() []types.Position { return nil }
func (s *stubConfigProvider) GetPlayer1Spawn() types.Position {
	return types.Position{}
}

func (s *stubConfigProvider) GetPlayer2Spawn() types.Position {
	return types.Position{}
}

func (s *stubConfigProvider) GetHQPosition() [2]int           { return [2]int{} }
func (s *stubConfigProvider) GetAIUpdateIntervalTicks() int   { return 0 }
func (s *stubConfigProvider) GetEnemyRespawnDelayTicks() uint { return 0 }

func (s *stubConfigProvider) GetBaseSizePx() uint { return s.baseSizePx }

func (s *stubConfigProvider) GetMapBlocksCount() types.Size { return types.Size{} }
func (s *stubConfigProvider) GetTileBaseSize() uint         { return 0 }
func (s *stubConfigProvider) GetTitleFontSize() uint        { return 0 }
func (s *stubConfigProvider) GetSubtitleFontSize() uint     { return 0 }
func (s *stubConfigProvider) GetRegularFontSize() uint      { return 0 }
func (s *stubConfigProvider) GetGameTitle() string          { return "" }
func (s *stubConfigProvider) GetVolume() float64            { return 0 }

type bonusTestEnv struct {
	tankCommon  *recordingTankCommon
	lifecycle   *recordingLifecycle
	session     *session_entities.StageSessionEntity
	bonusesRepo *game.BonusesRepository
	sounds      *use_cases.SoundUseCases
	registry    *testutil.FakeTilesetRegistry
	bonusUC     *use_cases.BonusUseCases
}

func newBonusTestEnv() *bonusTestEnv {
	tankCommon := &recordingTankCommon{}
	lifecycle := &recordingLifecycle{}
	session := session_entities.NewStageSessionEntity()
	bonusesRepo := game.NewBonusesRepository()
	sounds := use_cases.NewSoundUseCases()
	registry := &testutil.FakeTilesetRegistry{}
	tilesUC := use_cases.NewTilesUseCases(
		registry,
		types.TilesetTypeBonuses,
		nil,
		nil,
	)

	bonusUC := use_cases.NewBonusUseCases(
		tankCommon,
		lifecycle,
		session,
		bonusesRepo,
		&stubConfigProvider{baseSizePx: 16},
		tilesUC,
		&stubRenderUseCases{},
		sounds,
	)

	return &bonusTestEnv{
		tankCommon:  tankCommon,
		lifecycle:   lifecycle,
		session:     session,
		bonusesRepo: bonusesRepo,
		sounds:      sounds,
		registry:    registry,
		bonusUC:     bonusUC,
	}
}

func (env *bonusTestEnv) newBonus(
	bonusType types.BonusType,
) *types.BonusEntity {
	bonus := types.NewBonusEntity(
		bonusType,
		types.Position{X: 32, Y: 32},
		types.Size{Width: 16, Height: 16},
		nil,
	)
	env.bonusesRepo.AddBonus(bonus)
	return bonus
}

func newPlayerTank(role types.TankRole) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(role, types.DirectionUp)
	tank := &tankValue
	tank.State = types.TankStateStopped
	return tank
}

func newEnemyTankInState(state types.TankState) *types.TankEntity {
	tankValue := types.NewDefaultTankEntity(
		types.TankRoleEnemy,
		types.DirectionUp,
	)
	tank := &tankValue
	tank.State = state
	return tank
}

// Звезда: повышение уровня танка, бонус удаляется
func TestBonusUseCases_Apply_Star(t *testing.T) {
	env := newBonusTestEnv()
	bonus := env.newBonus(types.BonusTypeStar)
	tank := newPlayerTank(types.TankRolePlayer1)

	env.bonusUC.Apply(bonus, tank)

	if len(env.tankCommon.leveledUp) != 1 ||
		env.tankCommon.leveledUp[0] != tank {
		t.Errorf("LevelUp не вызван для танка: %v", env.tankCommon.leveledUp)
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 0 {
		t.Errorf("бонус не удалён: %d", got)
	}

	events := env.sounds.GetEvents()
	if len(events) != 1 || events[0].SoundID != types.SoundIDBonus {
		t.Errorf("ожидался звук бонуса, получено %v", events)
	}
}

// Граната: взрываются только активные враги
func TestBonusUseCases_Apply_Grenade(t *testing.T) {
	env := newBonusTestEnv()
	bonus := env.newBonus(types.BonusTypeGrenade)

	activeEnemy1 := newEnemyTankInState(types.TankStateStopped)
	activeEnemy2 := newEnemyTankInState(types.TankStateMoving)
	explodedEnemy := newEnemyTankInState(types.TankStateExploded)
	player := newPlayerTank(types.TankRolePlayer1)
	env.tankCommon.tanks = []*types.TankEntity{
		activeEnemy1,
		activeEnemy2,
		explodedEnemy,
		player,
		nil,
	}

	env.bonusUC.Apply(bonus, player)

	if len(env.lifecycle.exploded) != 2 {
		t.Fatalf(
			"ожидалось 2 взрыва, получено %d",
			len(env.lifecycle.exploded),
		)
	}
	if env.lifecycle.exploded[0] != activeEnemy1 ||
		env.lifecycle.exploded[1] != activeEnemy2 {
		t.Errorf("взорваны не те танки")
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 0 {
		t.Errorf("бонус не удалён: %d", got)
	}
}

// Танк: +1 жизнь игроку по роли танка
func TestBonusUseCases_Apply_Tank(t *testing.T) {
	env := newBonusTestEnv()
	env.session.SetPlayerCount(2)
	bonus := env.newBonus(types.BonusTypeTank)
	tank := newPlayerTank(types.TankRolePlayer2)

	env.bonusUC.Apply(bonus, tank)

	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer2); got != 4 {
		t.Errorf("жизни игрока 2: %d, ожидалось 4", got)
	}
	if got := env.session.GetPlayerLives(types.PlayerTankNumPlayer1); got != 3 {
		t.Errorf("жизни игрока 1 изменились: %d", got)
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 0 {
		t.Errorf("бонус не удалён: %d", got)
	}
}

// Каска/таймер/лопата пока не реализованы: бонус остаётся, но звук играет
func TestBonusUseCases_Apply_UnimplementedTypesKeepBonus(t *testing.T) {
	for _, bonusType := range []types.BonusType{
		types.BonusTypeHelmet,
		types.BonusTypeTimer,
		types.BonusTypeShovel,
	} {
		t.Run(string(bonusType), func(t *testing.T) {
			env := newBonusTestEnv()
			bonus := env.newBonus(bonusType)
			tank := newPlayerTank(types.TankRolePlayer1)

			env.bonusUC.Apply(bonus, tank)

			if got := len(env.bonusesRepo.GetAllBonuses()); got != 1 {
				t.Errorf("бонус удалён: осталось %d", got)
			}
			if len(env.tankCommon.leveledUp) != 0 {
				t.Errorf("неожиданный LevelUp")
			}
			if len(env.lifecycle.exploded) != 0 {
				t.Errorf("неожиданный взрыв")
			}
			if got := len(env.sounds.GetEvents()); got != 1 {
				t.Errorf("ожидался 1 звук, получено %d", got)
			}
		})
	}
}

func TestBonusUseCases_Apply_NilArguments(t *testing.T) {
	env := newBonusTestEnv()
	bonus := env.newBonus(types.BonusTypeStar)
	tank := newPlayerTank(types.TankRolePlayer1)

	env.bonusUC.Apply(nil, tank)
	env.bonusUC.Apply(bonus, nil)

	if got := len(env.sounds.GetEvents()); got != 0 {
		t.Errorf("звук запрошен при nil-аргументах: %d", got)
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 1 {
		t.Errorf("бонус удалён при nil-аргументах: %d", got)
	}
}

func TestBonusUseCases_GetRandomBonusType(t *testing.T) {
	env := newBonusTestEnv()
	allowed := map[types.BonusType]bool{
		types.BonusTypeGrenade: true,
		types.BonusTypeTank:    true,
		types.BonusTypeStar:    true,
	}

	for i := 0; i < 200; i++ {
		if got := env.bonusUC.GetRandomBonusType(); !allowed[got] {
			t.Fatalf("недопустимый тип бонуса: %q", got)
		}
	}
}

func TestBonusUseCases_SpawnRandomBonusEntity(t *testing.T) {
	env := newBonusTestEnv()
	position := types.Position{X: 48, Y: 96}
	allowed := map[types.BonusType]bool{
		types.BonusTypeGrenade: true,
		types.BonusTypeTank:    true,
		types.BonusTypeStar:    true,
	}

	bonus := env.bonusUC.SpawnRandomBonusEntity(position)
	if bonus == nil {
		t.Fatal("бонус не создан")
	}
	if !allowed[bonus.GetType()] {
		t.Errorf("недопустимый тип: %q", bonus.GetType())
	}
	if bonus.GetPosition() != position {
		t.Errorf("позиция %v, ожидалась %v", bonus.GetPosition(), position)
	}
	if bonus.GetSize() != (types.Size{Width: 16, Height: 16}) {
		t.Errorf("размер %v", bonus.GetSize())
	}

	// Изображение — статичный тайл с id типа бонуса
	id, err := bonus.GetImageID()
	if err != nil || id != string(bonus.GetType()) {
		t.Errorf("изображение бонуса: id=%q err=%v", id, err)
	}
}

func TestBonusUseCases_SpawnRandomBonusEntity_Failures(t *testing.T) {
	// Ошибка создания тайла -> nil
	env := newBonusTestEnv()
	env.registry.Err = errTileNotFound
	if got := env.bonusUC.SpawnRandomBonusEntity(types.Position{}); got != nil {
		t.Errorf("ожидался nil при ошибке тайла, получен %v", got)
	}
}
