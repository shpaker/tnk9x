package processed

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/types"
)

type SpritesRepository struct {
	fileRepo      raw.IFileRepository
	tilesetConfig types.SpritesConfig
	tilesetImage  *ebiten.Image
}

func NewSpritesRepository(fileRepo raw.IFileRepository) (*SpritesRepository, error) {
	// Загружаем конфиг спрайтов
	config, err := raw.ReadConfig[types.SpritesConfig](fileRepo.(*raw.FileRepository), filepath.Join("sprites"))
	if err != nil {
		return nil, err
	}

	// Загружаем изображение тайлсета
	img, err := fileRepo.ReadImage(filepath.Join("sprites"))
	if err != nil {
		return nil, err
	}
	tilesetImage := ebiten.NewImageFromImage(img)

	return &SpritesRepository{
		fileRepo:      fileRepo,
		tilesetConfig: config,
		tilesetImage:  tilesetImage,
	}, nil
}

func (sr *SpritesRepository) GetSprite(groupID string, spriteID string) (*ebiten.Image, error) {
	// Получаем группу спрайтов
	group, exists := sr.tilesetConfig[groupID]
	if !exists {
		return nil, fmt.Errorf("sprite group '%s' not found", groupID)
	}

	// Получаем данные о спрайте
	spriteData, exists := group[spriteID]
	if !exists {
		return nil, fmt.Errorf("sprite '%s' not found in group '%s'", spriteID, groupID)
	}

	// Вырезаем нужную область из тайлсета
	spriteImage := sr.tilesetImage.SubImage(
		image.Rectangle{
			Min: image.Point{X: spriteData.X, Y: spriteData.Y},
			Max: image.Point{X: spriteData.X + spriteData.W, Y: spriteData.Y + spriteData.H},
		},
	).(*ebiten.Image)

	return spriteImage, nil
}
