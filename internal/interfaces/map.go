package interfaces

import "github.com/shpaker/gonflict/internal/types"

// IMapObject определяет интерфейс для объектов карты
type IMapObject interface {
	GetSize() types.Size
	GetWorldPosition() types.Position
	GetScreenPosition() types.Position
}
