package stage

import (
	"github.com/shpaker/tnk9x/internal/types"
)

// RequiredSprites перечисляет спрайты HUD, запрашиваемые рендером
// боковой панели уровня
func RequiredSprites() types.SpriteManifest {
	return types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeHUD: {
				"enemy_icon",
				"life_icon",
				"roman_one",
				"roman_two",
				"letter_p",
				"flag_tl",
				"flag_tr",
				"flag_bl",
				"flag_br",
			},
		},
	}
}
