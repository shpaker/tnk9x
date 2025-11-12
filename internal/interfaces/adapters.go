package interfaces

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/types"
)

type IConfigProvider interface {
	GetEnemySpawners() []types.Position
	GetPlayer1Spawn() types.Position
	GetPlayer2Spawn() types.Position
	GetHQPosition() [2]int
	GetAIUpdateIntervalTicks() int
	GetEnemyRespawnDelayTicks() uint
	GetBaseSizePx() uint
	GetMapBlocksCount() types.Size
	GetMapOffsets() [2]uint
	GetTileBaseSize() uint
	GetTitleFontSize() uint
	GetSubtitleFontSize() uint
	GetRegularFontSize() uint
	GetGameTitle() string
}

type IInputAdapter interface {
	Update(dt float64)
}

type IAiInputAdapter interface {
	IInputAdapter
	AddTank(tank *types.TankEntity)
	RemoveTank(tank *types.TankEntity)
}

type ILuaEngine interface {
	Execute(script string) error

	CallFunction(functionName string, args ...lua.LValue) ([]lua.LValue, error)

	NewTable() *lua.LTable

	SetGlobal(name string, value lua.LValue)

	ToBool(value lua.LValue) bool

	ToNumber(value lua.LValue) lua.LNumber

	Close()
}

type IState interface {
	SetUp()
	Update()
	Draw(screen *ebiten.Image)
}
