package interfaces

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/models"
)

// IAssetsRepository определяет интерфейс для работы с ассетами
type IAssetsRepository interface {
	// ReadAsset читает файл и возвращает его содержимое в виде байтов
	ReadAsset(name string) ([]byte, error)

	// ReadImage читает изображение (добавляет расширение .png автоматически)
	ReadImage(name string) (image.Image, error)
}

// ISpritesRepository определяет интерфейс для работы со спрайтами
type ISpritesRepository interface {
	// GetSprite возвращает изображение спрайта по группе и идентификатору
	GetSprite(group_id string, sprite_id string) (*ebiten.Image, error)
}

// ILevelsService определяет интерфейс для работы с уровнями
type ILevelsService interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int) (models.Level, error)
}
