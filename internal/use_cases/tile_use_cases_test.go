package use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// routingRegistry записывает, через какой тайлсет запрошено изображение
type routingRegistry struct {
	*testutil.FakeTilesetRegistry
	playerIDs []string
	enemyIDs  []string
}

func (r *routingRegistry) GetPlayerImage(
	id string,
) (types.IImageProvider, error) {
	r.playerIDs = append(r.playerIDs, id)
	return r.FakeTilesetRegistry.GetPlayerImage(id)
}

func (r *routingRegistry) GetEnemyImage(
	id string,
) (types.IImageProvider, error) {
	r.enemyIDs = append(r.enemyIDs, id)
	return r.FakeTilesetRegistry.GetEnemyImage(id)
}

// countingAnimationService считает делегированные обновления анимаций
type countingAnimationService struct {
	updated []*image_providers.AnimationProvider
}

func (s *countingAnimationService) UpdateAnimation(
	animation *image_providers.AnimationProvider,
) {
	s.updated = append(s.updated, animation)
}

type tilesTestEnv struct {
	registry         *routingRegistry
	tileService      *testutil.FakeTileService
	animations       *game.AnimationsRepository
	animationService *countingAnimationService
	tilesUC          *use_cases.TilesUseCases
}

func newTilesTestEnv() *tilesTestEnv {
	registry := &routingRegistry{
		FakeTilesetRegistry: &testutil.FakeTilesetRegistry{},
	}
	tileService := &testutil.FakeTileService{}
	animations := game.NewAnimationsRepository()
	animationService := &countingAnimationService{}

	tilesUC := use_cases.NewTilesUseCasesWithAnimations(
		registry,
		types.TilesetTypeBlocks,
		animations,
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
		tileService,
		animationService,
	)

	return &tilesTestEnv{
		registry:         registry,
		tileService:      tileService,
		animations:       animations,
		animationService: animationService,
		tilesUC:          tilesUC,
	}
}

func TestTilesUseCases_GetImage(t *testing.T) {
	env := newTilesTestEnv()

	img, err := env.tilesUC.GetImage("brick")
	if err != nil || img == nil {
		t.Fatalf("изображение не получено: img=%v err=%v", img, err)
	}
	if len(env.registry.Requested) != 1 ||
		env.registry.Requested[0] != "brick" {
		t.Errorf("запрошены не те изображения: %v", env.registry.Requested)
	}

	env.registry.Err = errTileNotFound
	if _, err := env.tilesUC.GetImage("brick"); err == nil {
		t.Error("ожидалась ошибка реестра")
	}
}

// Танк игрока берётся из тайлсета player, вражеский — из enemy
func TestTilesUseCases_GetTankImage_Routing(t *testing.T) {
	env := newTilesTestEnv()

	if _, err := env.tilesUC.GetTankImage("tank_up", false); err != nil {
		t.Fatalf("изображение игрока: %v", err)
	}
	if _, err := env.tilesUC.GetTankImage("tank_down", true); err != nil {
		t.Fatalf("изображение врага: %v", err)
	}

	if len(env.registry.playerIDs) != 1 ||
		env.registry.playerIDs[0] != "tank_up" {
		t.Errorf("player-тайлсет: %v", env.registry.playerIDs)
	}
	if len(env.registry.enemyIDs) != 1 ||
		env.registry.enemyIDs[0] != "tank_down" {
		t.Errorf("enemy-тайлсет: %v", env.registry.enemyIDs)
	}
}

func TestTilesUseCases_CreateStaticTile(t *testing.T) {
	env := newTilesTestEnv()

	provider, err := env.tilesUC.CreateStaticTile("brick")
	if err != nil {
		t.Fatalf("статичный тайл: %v", err)
	}
	static, ok := provider.(*image_providers.StaticProvider)
	if !ok {
		t.Fatalf("провайдер не StaticProvider: %T", provider)
	}
	if static.ImageID != "brick" {
		t.Errorf("id тайла %q, ожидался brick", static.ImageID)
	}

	env.registry.Err = errTileNotFound
	if _, err := env.tilesUC.CreateStaticTile("brick"); err == nil {
		t.Error("ожидалась ошибка реестра")
	}
}

func TestTilesUseCases_CreateTankAnimationTile_Routing(t *testing.T) {
	env := newTilesTestEnv()

	if _, err := env.tilesUC.CreateTankAnimationTile("tank_up", false); err != nil {
		t.Fatalf("анимация игрока: %v", err)
	}
	if _, err := env.tilesUC.CreateTankAnimationTile("tank_up", true); err != nil {
		t.Fatalf("анимация врага: %v", err)
	}

	if len(env.tileService.Created) != 2 ||
		env.tileService.Created[0] != "player/tank_up" ||
		env.tileService.Created[1] != "enemy/tank_up" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}
}

func TestTilesUseCases_CreateSpawnAnimation(t *testing.T) {
	env := newTilesTestEnv()

	animation, err := env.tilesUC.CreateSpawnAnimation()
	if err != nil || animation == nil {
		t.Fatalf("анимация спавна: anim=%v err=%v", animation, err)
	}
	if len(env.tileService.Created) != 1 ||
		env.tileService.Created[0] != "spawner/spawner" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}

	// Анимация зарегистрирована в репозитории
	all := env.animations.GetAllAnimations()
	if len(all) != 1 || all[0] != animation {
		t.Errorf("анимация не зарегистрирована: %v", all)
	}
}

func TestTilesUseCases_CreateExplosionAnimation(t *testing.T) {
	env := newTilesTestEnv()

	animation, err := env.tilesUC.CreateExplosionAnimation()
	if err != nil || animation == nil {
		t.Fatalf("анимация взрыва: anim=%v err=%v", animation, err)
	}
	if len(env.tileService.Created) != 1 ||
		env.tileService.Created[0] != "explosion/explosion_tank" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}
	all := env.animations.GetAllAnimations()
	if len(all) != 1 || all[0] != animation {
		t.Errorf("анимация не зарегистрирована: %v", all)
	}
}

// Без специальных тайлсетов анимации спавна и взрыва недоступны
func TestTilesUseCases_CreateAnimations_EmptyTilesetType(t *testing.T) {
	env := newTilesTestEnv()
	tilesUC := use_cases.NewTilesUseCases(
		env.registry,
		types.TilesetTypeBlocks,
		env.tileService,
		env.animationService,
	)

	if _, err := tilesUC.CreateSpawnAnimation(); err == nil {
		t.Error("CreateSpawnAnimation: ожидалась ошибка")
	}
	if _, err := tilesUC.CreateExplosionAnimation(); err == nil {
		t.Error("CreateExplosionAnimation: ожидалась ошибка")
	}
}

func TestTilesUseCases_StartStopAnimation(t *testing.T) {
	env := newTilesTestEnv()
	animation := image_providers.NewAnimationProvider(
		types.AnimationData{{Image: "frame", Duration: 1}},
	)

	env.tilesUC.StartAnimation(animation)
	if !animation.IsAnimating {
		t.Error("анимация не запущена")
	}

	env.tilesUC.StopAnimation(animation)
	if animation.IsAnimating {
		t.Error("анимация не остановлена")
	}

	// nil-анимация не приводит к панике
	env.tilesUC.StartAnimation(nil)
	env.tilesUC.StopAnimation(nil)
}

// UpdateAnimations делегирует сервису все ненулевые анимации
func TestTilesUseCases_UpdateAnimations(t *testing.T) {
	env := newTilesTestEnv()
	first := image_providers.NewAnimationProvider(nil)
	second := image_providers.NewAnimationProvider(nil)
	env.tilesUC.AddAnimation(first)
	env.tilesUC.AddAnimation(nil)
	env.tilesUC.AddAnimation(second)

	env.tilesUC.UpdateAnimations()

	if len(env.animationService.updated) != 2 ||
		env.animationService.updated[0] != first ||
		env.animationService.updated[1] != second {
		t.Errorf("обновлённые анимации: %v", env.animationService.updated)
	}
}

// Без репозитория анимаций AddAnimation и UpdateAnimations — no-op
func TestTilesUseCases_NilAnimationsRepository(t *testing.T) {
	env := newTilesTestEnv()
	tilesUC := use_cases.NewTilesUseCases(
		env.registry,
		types.TilesetTypeBlocks,
		env.tileService,
		env.animationService,
	)

	tilesUC.AddAnimation(image_providers.NewAnimationProvider(nil))
	tilesUC.UpdateAnimations()

	if len(env.animationService.updated) != 0 {
		t.Errorf("неожиданные обновления: %v", env.animationService.updated)
	}
}
