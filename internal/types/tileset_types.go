package types

// TilesetType — категория тайлсета игровых спрайтов.
type TilesetType string

const (
	TilesetTypeBlocks          TilesetType = "blocks"
	TilesetTypePlayer          TilesetType = "player"
	TilesetTypeEnemy           TilesetType = "enemy"
	TilesetTypeBullet          TilesetType = "bullet"
	TilesetTypeSpawner         TilesetType = "spawner"
	TilesetTypeExplosion       TilesetType = "explosion"
	TilesetTypeBulletExplosion TilesetType = "bullet_explosion"
	TilesetTypeHQ              TilesetType = "hq"
	TilesetTypeBonuses         TilesetType = "bonuses"
	TilesetTypeHUD             TilesetType = "hud"
)

// AllTilesetTypes перечисляет все известные типы тайлсетов
func AllTilesetTypes() []TilesetType {
	return []TilesetType{
		TilesetTypeBlocks,
		TilesetTypePlayer,
		TilesetTypeEnemy,
		TilesetTypeBullet,
		TilesetTypeSpawner,
		TilesetTypeExplosion,
		TilesetTypeBulletExplosion,
		TilesetTypeHQ,
		TilesetTypeBonuses,
		TilesetTypeHUD,
	}
}

// TankTilesetType возвращает тайлсет танка по его принадлежности
func TankTilesetType(isEnemy bool) TilesetType {
	if isEnemy {
		return TilesetTypeEnemy
	}
	return TilesetTypePlayer
}
