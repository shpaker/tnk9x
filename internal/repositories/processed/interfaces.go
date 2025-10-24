package processed

import (
	"image"

	"github.com/shpaker/gonflict/internal/types"
)

// ITilesetRepository определяет интерфейс для работы с тайлсетами
type ITilesetRepository interface {
	// GetImage возвращает изображение по ID из тайлсета
	GetImage(id string) (image.Image, error)
	// GetAnimationData возвращает данные анимации по ID
	GetAnimationData(id string) (types.AnimationData, error)
}

// IMapsDataRepository определяет интерфейс для работы с картами уровней
type IMapsDataRepository interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int) ([]types.BlockEntity, error)
}
