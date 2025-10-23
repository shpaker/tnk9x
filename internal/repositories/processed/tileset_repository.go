package processed

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/types"
	"gopkg.in/yaml.v3"
)

type TilesetDataRepository struct {
	fileRepo       raw.IFileRepository
	imagesCache    map[string]*ebiten.Image
	animationsData map[string]types.AnimationData
}

func NewTilesetDataRepository(
	fileRepo raw.IFileRepository,
	tilesetName string,
) (*TilesetDataRepository, error) {
	// Загружаем конфигурацию тайлсета
	configData, err := fileRepo.ReadFile(filepath.Join("new", tilesetName) + ".yml")
	if err != nil {
		return nil, err
	}

	var config types.TilesetDataConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	// Загружаем изображение тайлсета
	img, err := fileRepo.ReadImage(filepath.Join("new", tilesetName))
	if err != nil {
		return nil, err
	}
	tilesetImage := ebiten.NewImageFromImage(img)

	// Создаем репозиторий
	repo := &TilesetDataRepository{
		fileRepo:       fileRepo,
		imagesCache:    make(map[string]*ebiten.Image),
		animationsData: make(map[string]types.AnimationData),
	}

	// Предварительно кэшируем все изображения
	tileSize := config.Size
	for imageID, coords := range config.Images {
		// Вырезаем нужную область из тайлсета
		spriteImage := tilesetImage.SubImage(
			image.Rectangle{
				Min: image.Point{X: coords[0] * tileSize, Y: coords[1] * tileSize},
				Max: image.Point{X: (coords[0] + 1) * tileSize, Y: (coords[1] + 1) * tileSize},
			},
		)

		// Создаем новое ebiten.Image и кэшируем его
		cachedImage := ebiten.NewImageFromImage(spriteImage)
		repo.imagesCache[imageID] = cachedImage
	}

	// Копируем данные анимаций
	for animationID, animation := range config.Animations {
		repo.animationsData[animationID] = animation
	}

	return repo, nil
}

func (tr *TilesetDataRepository) GetImage(
	id string,
) (*ebiten.Image, error) {
	// Проверяем кэш
	if cachedImage, exists := tr.imagesCache[id]; exists {
		return cachedImage, nil
	}

	// Если изображение не найдено в кэше, возвращаем ошибку
	return nil, fmt.Errorf("image '%s' not found", id)
}

func (tr *TilesetDataRepository) GetAnimationData(
	id string,
) (types.AnimationData, error) {
	// Получаем данные анимации
	animationData, exists := tr.animationsData[id]
	if !exists {
		return nil, fmt.Errorf("animation '%s' not found", id)
	}

	return animationData, nil
}
