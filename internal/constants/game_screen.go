package constants

const (
	TileMinSize             = 8
	BattleFieldBlocksLength = 26
	UpDownLeftPanelLength   = 2
	RightPanelLength        = 4
	BattleFieldOffset       = UpDownLeftPanelLength * TileMinSize
	BattleFieldWidthHeight  = BattleFieldBlocksLength * TileMinSize
	ScreenWidth             = BattleFieldWidthHeight + UpDownLeftPanelLength*TileMinSize + RightPanelLength*TileMinSize
	ScreenHeight            = BattleFieldWidthHeight + UpDownLeftPanelLength*TileMinSize*2
)
