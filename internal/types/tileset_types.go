package types

// TilesetType — категория тайлсета игровых спрайтов.
type TilesetType string

const (
	TilesetTypeBlocks    TilesetType = "blocks"
	TilesetTypePlayer    TilesetType = "player"
	TilesetTypeEnemy     TilesetType = "enemy"
	TilesetTypeBullet    TilesetType = "bullet"
	TilesetTypeSpawner   TilesetType = "spawner"
	TilesetTypeExplosion TilesetType = "explosion"
	TilesetTypeHQ        TilesetType = "hq"
	TilesetTypeBonuses   TilesetType = "bonuses"
	TilesetTypeHUD       TilesetType = "hud"
)
