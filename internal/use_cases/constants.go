package use_cases

// Игровые константы
const (
	TileMinSize           = 8
	TankSpriteSize        = 16
	MapBlocksLength       = 26
	MapWidthHeight        = MapBlocksLength * TileMinSize
	UpDownLeftPanelLength = 2
	RightPanelLength      = 4
	MapOffset             = UpDownLeftPanelLength * TileMinSize
	DT                    = 1.0 / 60.0
)
