package adapters

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
	"github.com/shpaker/gonflict/internal/utils"
)

// TilesAdapter адаптер для работы с тайлами
type TilesAdapter struct {
	tilesRepository processed.ITilesetRepository
	tileUseCases    *use_cases.TileUseCases
}

// NewTilesAdapter создает новый адаптер тайлов
func NewTilesAdapter(tilesRepository processed.ITilesetRepository) *TilesAdapter {
	return &TilesAdapter{
		tilesRepository: tilesRepository,
		tileUseCases:    use_cases.NewTileUseCases(tilesRepository),
	}
}

// GetTilesetRepository возвращает репозиторий тайлсетов
func (ta *TilesAdapter) GetTilesetRepository() processed.ITilesetRepository {
	return ta.tilesRepository
}

// GetTileUseCases возвращает use cases для работы с тайлами
func (ta *TilesAdapter) GetTileUseCases() interfaces.ITileUseCases {
	return ta.tileUseCases
}

// GetImage возвращает изображение блока
func (ta *TilesAdapter) GetImage(id *string, direction types.Direction) (*ebiten.Image, error) {
	if id == nil {
		return nil, fmt.Errorf("image ID is nil")
	}

	image, err := ta.tilesRepository.GetImage(*id)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	// Вычисляем угол поворота в зависимости от направления
	var angle float64
	switch direction {
	case types.DirectionUp:
		angle = 0
	case types.DirectionRight:
		angle = math.Pi / 2
	case types.DirectionDown:
		angle = math.Pi
	case types.DirectionLeft:
		angle = 3 * math.Pi / 2
	default:
		angle = 0
	}

	// Поворачиваем изображение в зависимости от направления
	return utils.RotateImageByAngle(image, angle)
}
