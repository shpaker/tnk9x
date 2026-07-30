package use_cases_test

import (
	"strings"
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

func newSpriteValidationManifest() types.SpriteManifest {
	return types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeHUD:    {"enemy_icon", "life_icon"},
			types.TilesetTypeBullet: {"bullet"},
		},
		Animations: map[types.TilesetType][]string{
			types.TilesetTypeSpawner: {"spawner"},
		},
	}
}

// Полный манифест против полного реестра — ошибок нет
func TestSpriteValidationUseCases_Validate_OK(t *testing.T) {
	registry := &testutil.FakeTilesetRegistry{}
	validationUC := use_cases.NewSpriteValidationUseCases(registry)

	if err := validationUC.Validate(newSpriteValidationManifest()); err != nil {
		t.Fatalf("неожиданная ошибка валидации: %v", err)
	}
}

// Все отсутствующие идентификаторы собираются в одну ошибку
func TestSpriteValidationUseCases_Validate_CollectsAllProblems(
	t *testing.T,
) {
	registry := &testutil.FakeTilesetRegistry{
		MissingIDs: []string{"hud/life_icon", "spawner/spawner"},
	}
	validationUC := use_cases.NewSpriteValidationUseCases(registry)

	err := validationUC.Validate(newSpriteValidationManifest())
	if err == nil {
		t.Fatal("ожидалась ошибка валидации")
	}

	message := err.Error()
	for _, expected := range []string{
		"image hud/life_icon",
		"animation spawner/spawner",
	} {
		if !strings.Contains(message, expected) {
			t.Errorf("в ошибке нет %q: %s", expected, message)
		}
	}
	if strings.Contains(message, "enemy_icon") {
		t.Errorf("существующий спрайт попал в ошибку: %s", message)
	}
}

// Пустой манифест валиден
func TestSpriteValidationUseCases_Validate_EmptyManifest(t *testing.T) {
	registry := &testutil.FakeTilesetRegistry{}
	validationUC := use_cases.NewSpriteValidationUseCases(registry)

	if err := validationUC.Validate(types.SpriteManifest{}); err != nil {
		t.Fatalf("пустой манифест должен быть валиден: %v", err)
	}
}

// Manifest.Merge объединяет списки обоих манифестов
func TestSpriteManifest_Merge(t *testing.T) {
	first := types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeHUD: {"enemy_icon"},
		},
	}
	second := types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeHUD:    {"life_icon"},
			types.TilesetTypeBullet: {"bullet"},
		},
		Animations: map[types.TilesetType][]string{
			types.TilesetTypeSpawner: {"spawner"},
		},
	}

	merged := first.Merge(second)

	hudIDs := merged.Images[types.TilesetTypeHUD]
	if len(hudIDs) != 2 || hudIDs[0] != "enemy_icon" ||
		hudIDs[1] != "life_icon" {
		t.Errorf("hud-изображения: %v", hudIDs)
	}
	if len(merged.Images[types.TilesetTypeBullet]) != 1 {
		t.Errorf("bullet-изображения: %v", merged.Images)
	}
	if len(merged.Animations[types.TilesetTypeSpawner]) != 1 {
		t.Errorf("анимации: %v", merged.Animations)
	}
}
