package repositories

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/types"
)

func TestGetSprite(t *testing.T) {
	// Создаем тестовое изображение 16x16
	testImage := image.NewRGBA(image.Rect(0, 0, 16, 16))

	// Создаем конфиг с тестовым спрайтом
	config := types.SpritesConfig{
		"base": {
			"test": types.SpriteData{X: 0, Y: 0, W: 8, H: 8},
		},
	}

	// Создаем SpritesRepository напрямую
	repo := &SpritesRepository{
		tilesetConfig: config,
		tilesetImage:  ebiten.NewImageFromImage(testImage),
	}

	// Вызываем GetSprite
	sprite, err := repo.GetSprite("base", "test")

	if err != nil {
		t.Errorf("неожиданная ошибка: %v", err)
		return
	}

	if sprite == nil {
		t.Error("ожидалось изображение спрайта, получен nil")
	}

	// Проверяем размер
	bounds := sprite.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf("ожидался размер 8x8, получен %dx%d", bounds.Dx(), bounds.Dy())
	}
}
