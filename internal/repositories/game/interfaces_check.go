package game

import "github.com/shpaker/gonflict/internal/interfaces"

// Проверка реализации интерфейсов на этапе компиляции
var (
	_ interfaces.IGameRepositoriesRegistry = (*GameRepositoriesRegistry)(nil)
	_ interfaces.ITanksRepository          = (*TanksRepository)(nil)
	_ interfaces.IBulletsRepository        = (*BulletsRepository)(nil)
	_ interfaces.IBlocksRepository         = (*BlocksRepository)(nil)
	_ interfaces.IAnimationsRepository     = (*AnimationsRepository)(nil)
)
