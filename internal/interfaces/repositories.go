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

// ILevelsDataService определяет интерфейс для работы с уровнями
type ILevelsDataService interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int) (models.Level, error)
}

// IPlayerService определяет интерфейс для работы с игроком
type IPlayerService interface {
	// GetPlayer возвращает данные игрока с начальными параметрами
	GetPlayer() (models.Tank, error)
	Update(dt float32)
}
