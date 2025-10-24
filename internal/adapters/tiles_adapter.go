package adapters

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// TilesAdapter адаптер для работы с тайлами
type TilesAdapter struct {
	tilesRepository types.ITilesetRepository
	tileUseCases    *use_cases.TileUseCases
}

// NewTilesAdapter создает новый адаптер тайлов
func NewTilesAdapter(tilesRepository types.ITilesetRepository) *TilesAdapter {
	return &TilesAdapter{
		tilesRepository: tilesRepository,
		tileUseCases:    use_cases.NewTileUseCases(tilesRepository),
	}
}

// GetTilesetRepository возвращает репозиторий тайлсетов
func (ta *TilesAdapter) GetTilesetRepository() types.ITilesetRepository {
	return ta.tilesRepository
}

// GetTileUseCases возвращает use cases для работы с тайлами
func (ta *TilesAdapter) GetTileUseCases() interfaces.ITileUseCases {
	return ta.tileUseCases
}
