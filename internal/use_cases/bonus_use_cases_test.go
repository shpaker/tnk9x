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

// stubConfigProvider отдаёт только размеры тайлов
type stubConfigProvider struct {
	baseSizePx   uint
	tileBaseSize uint
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

func (s *stubConfigProvider) GetTileBaseSize() uint     { return s.tileBaseSize }
func (s *stubConfigProvider) GetTitleFontSize() uint    { return 0 }
func (s *stubConfigProvider) GetSubtitleFontSize() uint { return 0 }
func (s *stubConfigProvider) GetRegularFontSize() uint  { return 0 }
func (s *stubConfigProvider) GetGameTitle() string      { return "" }
func (s *stubConfigProvider) GetVolume() float64        { return 0 }

type bonusTestEnv struct {
	tankCommon  *recordingTankCommon
	lifecycle   *recordingLifecycle
	session     *session_entities.StageSessionEntity
	mapEntity   *types.MapEntity
	bonusesRepo *game.BonusesRepository
	sounds      *use_cases.SoundUseCases
	registry    *testutil.FakeTilesetRegistry
	bonusUC     *use_cases.BonusUseCases
}

func newBonusTestEnv() *bonusTestEnv {
	tankCommon := &recordingTankCommon{}
	lifecycle := &recordingLifecycle{}
	session := session_entities.NewStageSessionEntity()
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		nil,
		nil,
	)
	bonusesRepo := game.NewBonusesRepository()
	sounds := use_cases.NewSoundUseCases(game.NewSoundEventsRepository())
	registry := &testutil.FakeTilesetRegistry{}
	tilesUC := use_cases.NewTilesUseCases(
		registry,
		types.TilesetTypeBonuses,
		nil,
		nil,
	)

	// Штаб 16x16 у нижней границы карты, как на реальных уровнях
	hq := &types.HQEntity{
		Position: types.Position{X: 96, Y: 192},
		Size:     types.Size{Width: 16, Height: 16},
	}

	bonusUC := use_cases.NewBonusUseCases(
		tankCommon,
		lifecycle,
		&stubHQUseCases{hq: hq},
		session,
		mapEntity,
		bonusesRepo,
		&stubConfigProvider{baseSizePx: 16, tileBaseSize: 8},
		tilesUC,
		&stubRenderUseCases{},
		sounds,
	)

	return &bonusTestEnv{
		tankCommon:  tankCommon,
		lifecycle:   lifecycle,
		session:     session,
		mapEntity:   mapEntity,
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

// Каска: танк получает щит, бонус удаляется; щит истекает по отсчёту
func TestBonusUseCases_Apply_Helmet(t *testing.T) {
	env := newBonusTestEnv()
	bonus := env.newBonus(types.BonusTypeHelmet)
	tank := newPlayerTank(types.TankRolePlayer1)
	env.tankCommon.tanks = []*types.TankEntity{tank}

	env.bonusUC.Apply(bonus, tank)

	if !tank.HasShield() {
		t.Fatal("щит не активирован")
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 0 {
		t.Errorf("бонус не удалён: %d", got)
	}

	// 10 секунд по 60 тиков — щит истекает
	for i := 0; i < 10*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	if tank.HasShield() {
		t.Error("щит не истёк")
	}
}

// Таймер: враги замораживаются, бонус удаляется; заморозка истекает
func TestBonusUseCases_Apply_Timer(t *testing.T) {
	env := newBonusTestEnv()
	bonus := env.newBonus(types.BonusTypeTimer)
	tank := newPlayerTank(types.TankRolePlayer1)

	env.bonusUC.Apply(bonus, tank)

	if !env.session.AreEnemiesFrozen() {
		t.Fatal("враги не заморожены")
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 0 {
		t.Errorf("бонус не удалён: %d", got)
	}

	for i := 0; i < 10*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	if env.session.AreEnemiesFrozen() {
		t.Error("заморозка не истекла")
	}
}

// Лопата: кольцо вокруг штаба становится бетонным, бонус удаляется;
// по истечении отсчёта возвращаются исходные блоки
func TestBonusUseCases_Apply_Shovel(t *testing.T) {
	env := newBonusTestEnv()

	// Кирпич в кольце штаба и кирпич в стороне от него
	wallBrick := types.NewBlockEntity(string(types.Brick), 88, 184, 8, nil)
	farBrick := types.NewBlockEntity(string(types.Brick), 0, 0, 8, nil)
	env.mapEntity.AddBlock(wallBrick)
	env.mapEntity.AddBlock(farBrick)

	bonus := env.newBonus(types.BonusTypeShovel)
	tank := newPlayerTank(types.TankRolePlayer1)

	env.bonusUC.Apply(bonus, tank)

	if !env.mapEntity.IsHQFortified() {
		t.Fatal("укрепление не активировано")
	}
	if got := len(env.bonusesRepo.GetAllBonuses()); got != 0 {
		t.Errorf("бонус не удалён: %d", got)
	}

	// Штаб (96,192) 16x16 у нижней границы карты 208x208:
	// кольцо из 8 клеток по 8px, нижний ряд за границей
	steelCount := 0
	for _, block := range env.mapEntity.GetBlocks() {
		if block.Data != nil && block.Data.Name == types.Steel {
			steelCount++
		}
	}
	if steelCount != 8 {
		t.Errorf("бетонных блоков %d, ожидалось 8", steelCount)
	}

	// Кирпич из кольца снят с карты, дальний остался
	blocksSet := map[*types.BlockEntity]bool{}
	for _, block := range env.mapEntity.GetBlocks() {
		blocksSet[block] = true
	}
	if blocksSet[wallBrick] {
		t.Error("кирпич кольца не заменён бетоном")
	}
	if !blocksSet[farBrick] {
		t.Error("дальний кирпич пропал")
	}

	// По истечении отсчёта бетон снят, кирпич кольца возвращён
	for i := 0; i < 20*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	if env.mapEntity.IsHQFortified() {
		t.Fatal("укрепление не истекло")
	}
	steelCount = 0
	restored := false
	for _, block := range env.mapEntity.GetBlocks() {
		if block.Data != nil && block.Data.Name == types.Steel {
			steelCount++
		}
		if block == wallBrick {
			restored = true
		}
	}
	if steelCount != 0 {
		t.Errorf("бетон не снят: %d блоков", steelCount)
	}
	if !restored {
		t.Error("кирпич кольца не восстановлен")
	}
}

// Повторная лопата продлевает укрепление, не дублируя бетон
func TestBonusUseCases_Apply_Shovel_Prolong(t *testing.T) {
	env := newBonusTestEnv()
	tank := newPlayerTank(types.TankRolePlayer1)

	env.bonusUC.Apply(env.newBonus(types.BonusTypeShovel), tank)

	// Половина отсчёта прошла, затем вторая лопата
	for i := 0; i < 10*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	env.bonusUC.Apply(env.newBonus(types.BonusTypeShovel), tank)

	steelCount := 0
	for _, block := range env.mapEntity.GetBlocks() {
		if block.Data != nil && block.Data.Name == types.Steel {
			steelCount++
		}
	}
	if steelCount != 8 {
		t.Errorf("бетонных блоков %d, ожидалось 8", steelCount)
	}

	// Отсчёт начат заново: спустя ещё половину срока укрепление активно
	for i := 0; i < 10*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	if !env.mapEntity.IsHQFortified() {
		t.Error("укрепление истекло раньше продлённого срока")
	}
	for i := 0; i < 10*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	if env.mapEntity.IsHQFortified() {
		t.Error("продлённое укрепление не истекло")
	}
}

// Лопата с уже сколотым пулями кирпичом в кольце: остаток заменяется
// бетоном и восстанавливается в прежнем виде по истечении отсчёта
func TestBonusUseCases_Apply_Shovel_RemnantInRing(t *testing.T) {
	env := newBonusTestEnv()

	// Остаток 8x4 в клетке кольца (88,184): тайл сколот сверху
	remnant := types.NewBlockEntity(string(types.Brick), 88, 184, 8, nil)
	remnant.Position = types.Position{X: 88, Y: 188}
	remnant.Size = types.Size{Width: 8, Height: 4}
	env.mapEntity.AddBlock(remnant)

	env.bonusUC.Apply(
		env.newBonus(types.BonusTypeShovel),
		newPlayerTank(types.TankRolePlayer1),
	)

	steelCount := 0
	for _, block := range env.mapEntity.GetBlocks() {
		if block == remnant {
			t.Fatal("остаток кирпича выжил под бетоном")
		}
		if block.Data != nil && block.Data.Name == types.Steel {
			steelCount++
		}
	}
	if steelCount != 8 {
		t.Errorf("бетонных блоков %d, ожидалось 8", steelCount)
	}

	// По истечении отсчёта остаток возвращается тем же прямоугольником
	for i := 0; i < 20*60; i++ {
		env.bonusUC.UpdateEffects()
	}
	restored := false
	for _, block := range env.mapEntity.GetBlocks() {
		if block == remnant {
			restored = true
		}
	}
	if !restored {
		t.Fatal("остаток кирпича не восстановлен")
	}
	if remnant.Position != (types.Position{X: 88, Y: 188}) ||
		remnant.Size != (types.Size{Width: 8, Height: 4}) {
		t.Errorf(
			"остаток изменился: %v %dx%d",
			remnant.Position,
			remnant.Size.Width,
			remnant.Size.Height,
		)
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
		types.BonusTypeHelmet:  true,
		types.BonusTypeTimer:   true,
		types.BonusTypeShovel:  true,
		types.BonusTypeGrenade: true,
		types.BonusTypeTank:    true,
		types.BonusTypeStar:    true,
	}

	seen := map[types.BonusType]bool{}
	for i := 0; i < 600; i++ {
		got := env.bonusUC.GetRandomBonusType()
		if !allowed[got] {
			t.Fatalf("недопустимый тип бонуса: %q", got)
		}
		seen[got] = true
	}
	if len(seen) != len(allowed) {
		t.Errorf("выпали не все типы бонусов: %v", seen)
	}
}

func TestBonusUseCases_SpawnRandomBonusEntity(t *testing.T) {
	env := newBonusTestEnv()
	position := types.Position{X: 48, Y: 96}
	allowed := map[types.BonusType]bool{
		types.BonusTypeHelmet:  true,
		types.BonusTypeTimer:   true,
		types.BonusTypeShovel:  true,
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

// VisibleBonuses отдаёт только бонусы в видимой фазе мигания
func TestBonusUseCases_VisibleBonuses(t *testing.T) {
	env := newBonusTestEnv()

	visibleBonus := types.NewBonusEntity(
		types.BonusTypeStar,
		types.Position{},
		types.Size{Width: 16, Height: 16},
		nil,
	)
	hiddenBonus := types.NewBonusEntity(
		types.BonusTypeGrenade,
		types.Position{},
		types.Size{Width: 16, Height: 16},
		nil,
	)
	for i := 0; i < 10; i++ {
		hiddenBonus.UpdateBlink()
	}

	env.bonusesRepo.AddBonus(visibleBonus)
	env.bonusesRepo.AddBonus(nil)
	env.bonusesRepo.AddBonus(hiddenBonus)

	visible := env.bonusUC.VisibleBonuses()
	if len(visible) != 1 || visible[0] != visibleBonus {
		t.Errorf("видимые бонусы: %v", visible)
	}
}

// Пустой репозиторий — пустой список видимых бонусов
func TestBonusUseCases_VisibleBonuses_Empty(t *testing.T) {
	env := newBonusTestEnv()

	if visible := env.bonusUC.VisibleBonuses(); len(visible) != 0 {
		t.Errorf("ожидался пустой список, получено: %v", visible)
	}
}
