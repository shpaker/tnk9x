package interfaces

import lua "github.com/yuin/gopher-lua"

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
