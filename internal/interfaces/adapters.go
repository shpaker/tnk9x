package interfaces

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/types"
)

// IConfigProvider интерфейс для получения данных конфигурации
type IConfigProvider interface {
	GetEnemySpawners() []types.Position
	GetPlayer1Spawn() types.Position
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

// IInputAdapter интерфейс для адаптеров ввода
type IInputAdapter interface {
	Update(dt float64)
}

// IAiInputAdapter интерфейс для AI адаптера танков
type IAiInputAdapter interface {
	IInputAdapter
	AddTank(tank *types.TankEntity)
	RemoveTank(tank *types.TankEntity)
}

// ILuaEngine интерфейс для работы с Lua VM (Infrastructure Layer)
type ILuaEngine interface {
	// Execute выполняет Lua скрипт из строки
	Execute(script string) error

	// CallFunction вызывает Lua функцию с параметрами и возвращает результаты
	CallFunction(functionName string, args ...lua.LValue) ([]lua.LValue, error)

	// NewTable создает новую Lua таблицу
	NewTable() *lua.LTable

	// SetGlobal устанавливает глобальную переменную в Lua
	SetGlobal(name string, value lua.LValue)

	// ToBool конвертирует Lua значение в bool
	ToBool(value lua.LValue) bool

	// ToNumber конвертирует Lua значение в число
	ToNumber(value lua.LValue) lua.LNumber

	// Close освобождает ресурсы Lua VM
	Close()
}

// IState интерфейс для состояний игры
type IState interface {
	SetUp()
	Update()
	Draw(screen *ebiten.Image)
}
