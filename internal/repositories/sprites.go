package repositories

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

type SpritesRepository struct {
	assetsRepository *interfaces.IAssetsRepository
	tilesetConfig    types.SpritesConfig
	tilesetImage     *ebiten.Image
}

func NewSpritesRepository(assetsRepository interfaces.IAssetsRepository) (*SpritesRepository, error) {
	// Загружаем конфиг спрайтов
	config, err := ReadConfig[types.SpritesConfig](assetsRepository, filepath.Join("sprites"))
	if err != nil {
		return nil, err
	}

	// Загружаем изображение тайлсета
	img, err := assetsRepository.ReadImage(filepath.Join("sprites"))
	if err != nil {
		return nil, err
	}
	tilesetImage := ebiten.NewImageFromImage(img)

	return &SpritesRepository{
		assetsRepository: &assetsRepository,
		tilesetConfig:    config,
		tilesetImage:     tilesetImage,
	}, nil
}

func (r *SpritesRepository) GetSprite(group_id string, sprite_id string) (*ebiten.Image, error) {
	// Получаем группу спрайтов
	group, exists := r.tilesetConfig[group_id]
	if !exists {
		return nil, fmt.Errorf("sprite group '%s' not found", group_id)
	}

	// Получаем данные о спрайте
	spriteData, exists := group[sprite_id]
	if !exists {
		return nil, fmt.Errorf("sprite '%s' not found in group '%s'", sprite_id, group_id)
	}

	// Вырезаем нужную область из тайлсета
	spriteImage := r.tilesetImage.SubImage(
		image.Rectangle{
			Min: image.Point{X: spriteData.X, Y: spriteData.Y},
			Max: image.Point{X: spriteData.X + spriteData.W, Y: spriteData.Y + spriteData.H},
		},
	).(*ebiten.Image)

	return spriteImage, nil
}
