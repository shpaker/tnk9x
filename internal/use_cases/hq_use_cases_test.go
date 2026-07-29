package use_cases_test

import (
	"testing"

	game "github.com/shpaker/tnk9x/internal/repositories/game"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type hqTestEnv struct {
	registry    *testutil.FakeTilesetRegistry
	tileService *testutil.FakeTileService
	animations  *game.AnimationsRepository
	hq          *types.HQEntity
	hqUC        *use_cases.HQUseCases
}

func newHQTestEnv() *hqTestEnv {
	registry := &testutil.FakeTilesetRegistry{}
	tileService := &testutil.FakeTileService{}
	animations := game.NewAnimationsRepository()
	tilesUC := use_cases.NewTilesUseCasesWithAnimations(
		registry,
		types.TilesetTypeHQ,
		animations,
		types.TilesetTypeSpawner,
		types.TilesetTypeExplosion,
		tileService,
		nil,
	)

	hq := &types.HQEntity{
		Position: types.Position{X: 96, Y: 192},
		Size:     types.Size{Width: 16, Height: 16},
		Altitude: types.SURFACE,
		Image:    &stubImageProvider{},
		State:    types.HQStateIntact,
	}

	return &hqTestEnv{
		registry:    registry,
		tileService: tileService,
		animations:  animations,
		hq:          hq,
		hqUC:        use_cases.NewHQUseCases(tilesUC, hq),
	}
}

// Взрыв целого штаба: анимация взрыва запускается вместо изображения
func TestHQUseCases_Explode(t *testing.T) {
	env := newHQTestEnv()
	intactImage := env.hq.Image

	if err := env.hqUC.Explode(env.hq); err != nil {
		t.Fatalf("взрыв штаба: %v", err)
	}

	if env.hq.State != types.HQStateExploding {
		t.Errorf("состояние %v, ожидалось Exploding", env.hq.State)
	}
	if env.hq.Image == intactImage {
		t.Error("изображение не заменено анимацией взрыва")
	}
	animation, ok := env.hq.Image.(*image_providers.AnimationProvider)
	if !ok {
		t.Fatalf("изображение не AnimationProvider: %T", env.hq.Image)
	}
	if !animation.IsAnimating {
		t.Error("анимация взрыва не запущена")
	}
	if len(env.tileService.Created) != 1 ||
		env.tileService.Created[0] != "explosion/explosion_tank" {
		t.Errorf("созданные анимации: %v", env.tileService.Created)
	}
	if got := len(env.animations.GetAllAnimations()); got != 1 {
		t.Errorf("анимаций в репозитории %d, ожидалась 1", got)
	}
}

// Повторный взрыв и взрыв разрушенного штаба — no-op
func TestHQUseCases_Explode_NoOp(t *testing.T) {
	for _, state := range []types.HQState{
		types.HQStateExploding,
		types.HQStateDestroyed,
	} {
		env := newHQTestEnv()
		env.hq.State = state
		image := env.hq.Image

		if err := env.hqUC.Explode(env.hq); err != nil {
			t.Fatalf("взрыв в состоянии %v: %v", state, err)
		}
		if env.hq.State != state || env.hq.Image != image {
			t.Errorf("состояние %v изменено", state)
		}
		if len(env.tileService.Created) != 0 {
			t.Errorf("создана лишняя анимация: %v", env.tileService.Created)
		}
	}

	env := newHQTestEnv()
	if err := env.hqUC.Explode(nil); err != nil {
		t.Errorf("взрыв nil-штаба: %v", err)
	}
}

// Ошибка тайл-сервиса прерывает взрыв без смены состояния
func TestHQUseCases_Explode_TileError(t *testing.T) {
	env := newHQTestEnv()
	env.tileService.Err = errTileNotFound
	intactImage := env.hq.Image

	if err := env.hqUC.Explode(env.hq); err == nil {
		t.Error("ожидалась ошибка взрыва")
	}
	if env.hq.State != types.HQStateIntact || env.hq.Image != intactImage {
		t.Error("состояние изменено при ошибке")
	}
}

// Завершение анимации взрыва: статичный тайл руин, состояние Destroyed
func TestHQUseCases_IsExplosionFinished(t *testing.T) {
	env := newHQTestEnv()
	if err := env.hqUC.Explode(env.hq); err != nil {
		t.Fatalf("взрыв штаба: %v", err)
	}
	animation := env.hq.Image.(*image_providers.AnimationProvider)

	// Анимация ещё идёт — состояние не меняется
	env.hqUC.IsExplosionFinished(env.hq)
	if env.hq.State != types.HQStateExploding {
		t.Fatalf("состояние %v, ожидалось Exploding", env.hq.State)
	}

	animation.IsAnimating = false
	env.hqUC.IsExplosionFinished(env.hq)

	if env.hq.State != types.HQStateDestroyed {
		t.Errorf("состояние %v, ожидалось Destroyed", env.hq.State)
	}
	if env.hq.Altitude != types.SURFACE {
		t.Errorf("высота %v, ожидалась SURFACE", env.hq.Altitude)
	}
	static, ok := env.hq.Image.(*image_providers.StaticProvider)
	if !ok {
		t.Fatalf("изображение не StaticProvider: %T", env.hq.Image)
	}
	if static.ImageID != "hq_destroyed" {
		t.Errorf("id тайла %q, ожидался hq_destroyed", static.ImageID)
	}
	if !env.hqUC.IsDestroyed() {
		t.Error("IsDestroyed: ожидалось true")
	}
}

// Ошибка тайла руин: штаб всё равно разрушается, изображение остаётся
func TestHQUseCases_IsExplosionFinished_TileError(t *testing.T) {
	env := newHQTestEnv()
	if err := env.hqUC.Explode(env.hq); err != nil {
		t.Fatalf("взрыв штаба: %v", err)
	}
	animation := env.hq.Image.(*image_providers.AnimationProvider)
	animation.IsAnimating = false
	env.registry.Err = errTileNotFound

	env.hqUC.IsExplosionFinished(env.hq)

	if env.hq.State != types.HQStateDestroyed {
		t.Errorf("состояние %v, ожидалось Destroyed", env.hq.State)
	}
	if env.hq.Image != animation {
		t.Errorf("изображение заменено при ошибке: %T", env.hq.Image)
	}
}

// Вне состояния Exploding завершение взрыва не проверяется
func TestHQUseCases_IsExplosionFinished_NoOp(t *testing.T) {
	env := newHQTestEnv()
	image := env.hq.Image

	env.hqUC.IsExplosionFinished(env.hq)
	env.hqUC.IsExplosionFinished(nil)

	if env.hq.State != types.HQStateIntact || env.hq.Image != image {
		t.Error("состояние целого штаба изменено")
	}
}

func TestHQUseCases_GetHQAndIsDestroyed(t *testing.T) {
	env := newHQTestEnv()

	if got := env.hqUC.GetHQ(); got != env.hq {
		t.Errorf("GetHQ вернул %v", got)
	}
	if env.hqUC.IsDestroyed() {
		t.Error("целый штаб считается разрушенным")
	}

	env.hq.State = types.HQStateDestroyed
	if !env.hqUC.IsDestroyed() {
		t.Error("разрушенный штаб не распознан")
	}

	nilUC := use_cases.NewHQUseCases(nil, nil)
	if nilUC.IsDestroyed() {
		t.Error("nil-штаб считается разрушенным")
	}
}
