package interfaces

import "github.com/hajimehoshi/ebiten/v2"

// ISpritesRepository определяет интерфейс для работы со спрайтами
type ISpritesRepository interface {
	// GetSprite возвращает изображение спрайта по группе и идентификатору
	GetSprite(group_id string, sprite_id string) (*ebiten.Image, error)
}
