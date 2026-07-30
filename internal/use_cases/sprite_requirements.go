package use_cases

import (
	"github.com/shpaker/tnk9x/internal/types"
)

// RequiredSprites перечисляет спрайты и анимации, запрашиваемые
// use case'ами: пуля, штаб, бонусы и анимации танков, спавна и взрыва
func RequiredSprites() types.SpriteManifest {
	return types.SpriteManifest{
		Images: map[types.TilesetType][]string{
			types.TilesetTypeBullet:  {"bullet"},
			types.TilesetTypeHQ:      {"hq_destroyed"},
			types.TilesetTypeBonuses: bonusSpriteIDs(),
		},
		Animations: map[types.TilesetType][]string{
			types.TilesetTypeSpawner:         {"spawner"},
			types.TilesetTypeExplosion:       {"explosion_tank"},
			types.TilesetTypeBulletExplosion: {"bullet_explosion"},
			types.TilesetTypePlayer: tankAnimationIDs(
				types.TankRolePlayer1,
				types.TankRolePlayer2,
			),
			types.TilesetTypeEnemy: tankAnimationIDs(types.TankRoleEnemy),
		},
	}
}

// bonusSpriteIDs — все типы бонусов рисуются одноимёнными спрайтами
func bonusSpriteIDs() []string {
	bonusTypes := []types.BonusType{
		types.BonusTypeHelmet,
		types.BonusTypeTimer,
		types.BonusTypeShovel,
		types.BonusTypeStar,
		types.BonusTypeGrenade,
		types.BonusTypeTank,
	}

	ids := make([]string, 0, len(bonusTypes))
	for _, bonusType := range bonusTypes {
		ids = append(ids, string(bonusType))
	}
	return ids
}

// tankAnimationIDs разворачивает матрицу анимаций танков:
// роли x модели 1-4 x направления, формат имени общий с TankEntity
func tankAnimationIDs(roles ...types.TankRole) []string {
	var ids []string
	for _, role := range roles {
		for model := uint(1); model <= 4; model++ {
			for _, direction := range types.TankAnimationDirections() {
				ids = append(
					ids,
					types.TankAnimationNameFor(role, model, direction),
				)
			}
		}
	}
	return ids
}
