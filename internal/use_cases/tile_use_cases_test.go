package use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

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
	registry         *testutil.FakeTilesetRegistry
	tileService      *testutil.FakeTileService
	animations       *game.AnimationsRepository
	animationService *countingAnimationService
	tilesUC          *use_cases.TilesUseCases
}

func newTilesTestEnv() *tilesTestEnv {
	registry := &testutil.FakeTilesetRegistry{}
	tileService := &testutil.FakeTileService{}
	animations := game.NewAnimationsRepository()
	animationService := &countingAnimationService{}

	tilesUC := use_cases.NewTilesUseCasesWithAnimations(
		registry,
		types.TilesetTypeBlocks,
		animations,
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

// Без репозитория анимаций анимации спавна и взрыва недоступны
func TestTilesUseCases_CreateAnimations_WithoutAnimationsRepo(t *testing.T) {
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
