package processed

import "github.com/shpaker/gonflict/internal/interfaces"

// Проверка реализации интерфейсов на этапе компиляции
var (
	_ interfaces.ITilesetRepository         = (*TilesetDataRepository)(nil)
	_ interfaces.IMapsDataRepository        = (*MapsDataRepository)(nil)
	_ interfaces.ITilesetRepositoryRegistry = (*TilesetRepositoryRegistry)(nil)
	_ interfaces.IScriptsRepository         = (*ScriptsRepository)(nil)
)
