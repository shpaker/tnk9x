package interfaces

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/types"
)

// IConfigProvider интерфейс для получения данных конфигурации
type IConfigProvider interface {
	GetEnemySpawners() [][2]int
	GetPlayerSpawners() [][2]int
	GetHQPosition() [2]int
	GetAIUpdateIntervalTicks() int
	GetBaseSizePx() uint
	GetMapBlocksCount() types.Size
	GetMapOffsets() [2]uint
	GetTileBaseSize() uint
}

// IInputAdapter интерфейс для адаптеров ввода
type IInputAdapter interface {
	Update(dt float64)
}

// ILuaEngine интерфейс для работы с Lua VM (Infrastructure Layer)
type ILuaEngine interface {
	// Execute выполняет Lua скрипт из строки
	Execute(script string) error

	// CallFunction вызывает Lua функцию с параметрами и возвращает результаты
	CallFunction(functionName string, args ...lua.LValue) ([]lua.LValue, error)

	// NewTable создает новую Lua таблицу
	NewTable() *lua.LTable

	// ToBool конвертирует Lua значение в bool
	ToBool(value lua.LValue) bool

	// ToNumber конвертирует Lua значение в число
	ToNumber(value lua.LValue) lua.LNumber

	// Close освобождает ресурсы Lua VM
	Close()
}
