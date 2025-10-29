package input_adapters

import "github.com/shpaker/gonflict/internal/interfaces"

// Проверка реализации интерфейсов на этапе компиляции
var (
	_ interfaces.IInputAdapter = (*KeyboardInputAdapter)(nil)
	_ interfaces.IInputAdapter = (*AiInputAdapter)(nil)
)
