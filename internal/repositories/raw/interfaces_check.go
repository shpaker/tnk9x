package raw

import "github.com/shpaker/gonflict/internal/interfaces"

// Проверка реализации интерфейсов на этапе компиляции
var _ interfaces.IFileRepository = (*FileRepository)(nil)
