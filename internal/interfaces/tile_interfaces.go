package interfaces

import (
	"github.com/shpaker/gonflict/internal/types"
)

// ITileUseCases определяет интерфейс для работы с тайлами
type ITileUseCases interface {
	CreateStaticTile(id string) (types.IImageIdGetter, error)
	CreateAnimationTile(id string) (*types.TileAnimationEntity, error)
}
