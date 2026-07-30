package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// Изображение запрашивается из тайлсета, указанного вызывающим
func TestSpriteUseCases_GetImage_Routing(t *testing.T) {
	registry := &testutil.FakeTilesetRegistry{}
	spriteUC := use_cases.NewSpriteUseCases(registry)

	tests := []struct {
		name        string
		tilesetType types.TilesetType
		imageID     string
		requested   string
	}{
		{"блок карты", types.TilesetTypeBlocks, "brick", "blocks/brick"},
		{"иконка HUD", types.TilesetTypeHUD, "enemy_icon", "hud/enemy_icon"},
		{
			"танк игрока",
			types.TankTilesetType(false),
			"tank_up",
			"player/tank_up",
		},
		{
			"танк врага",
			types.TankTilesetType(true),
			"tank_down",
			"enemy/tank_down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry.Requested = nil

			img, err := spriteUC.GetImage(tt.tilesetType, tt.imageID)
			if err != nil || img == nil {
				t.Fatalf("изображение не получено: img=%v err=%v", img, err)
			}
			if len(registry.Requested) != 1 ||
				registry.Requested[0] != tt.requested {
				t.Errorf(
					"запрошено %v, ожидалось [%s]",
					registry.Requested,
					tt.requested,
				)
			}
		})
	}
}

func TestSpriteUseCases_GetImage_RegistryError(t *testing.T) {
	registry := &testutil.FakeTilesetRegistry{Err: errTileNotFound}
	spriteUC := use_cases.NewSpriteUseCases(registry)

	if _, err := spriteUC.GetImage(
		types.TilesetTypeBlocks,
		"brick",
	); err == nil {
		t.Error("ожидалась ошибка реестра")
	}
}

// GetImageIDs отдаёт список спрайтов тайлсета как есть
func TestSpriteUseCases_GetImageIDs(t *testing.T) {
	registry := &testutil.FakeTilesetRegistry{
		ImageIDs: []string{"brick", "steel"},
	}
	spriteUC := use_cases.NewSpriteUseCases(registry)

	ids := spriteUC.GetImageIDs(types.TilesetTypeBlocks)
	if len(ids) != 2 || ids[0] != "brick" || ids[1] != "steel" {
		t.Errorf("список спрайтов: %v", ids)
	}
}
