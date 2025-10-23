package processed

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/types"
)

// ISpritesRepository определяет интерфейс для работы со спрайтами
type ISpritesRepository interface {
	// GetSprite возвращает изображение спрайта по группе и идентификатору
	GetSprite(groupID string, spriteID string) (*ebiten.Image, error)
}

// IMapsDataRepository определяет интерфейс для работы с картами уровней
type IMapsDataRepository interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int) ([]types.Block, error)
}
