package processed

import (
	"fmt"
	"image"

	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/types"
	"gopkg.in/yaml.v3"
)

type TilesetDataRepository struct {
	fileRepo         raw.IFileRepository
	imagesCache      map[string]image.Image
	animationsData   map[string]types.AnimationData
	animationsConfig map[string]types.AnimationConfig // Храним полную конфигурацию
}

func NewTilesetDataRepository(
	fileRepo raw.IFileRepository,
	tilesetName string,
) (*TilesetDataRepository, error) {
	// Загружаем конфигурацию тайлсета
	configData, err := fileRepo.ReadFile(tilesetName + ".yml")
	if err != nil {
		return nil, err
	}

	var config types.TilesetDataConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	// Загружаем изображение тайлсета
	img, err := fileRepo.ReadImage(tilesetName)
	if err != nil {
		return nil, err
	}

	// Создаем репозиторий
	repo := &TilesetDataRepository{
		fileRepo:         fileRepo,
		imagesCache:      make(map[string]image.Image),
		animationsData:   make(map[string]types.AnimationData),
		animationsConfig: make(map[string]types.AnimationConfig),
	}

	// Предварительно кэшируем все изображения
	tileSize := config.Size
	for imageID, coords := range config.Images {
		// Вырезаем нужную область из тайлсета
		spriteImage := img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(
			image.Rectangle{
				Min: image.Point{X: coords[0] * tileSize, Y: coords[1] * tileSize},
				Max: image.Point{X: (coords[0] + 1) * tileSize, Y: (coords[1] + 1) * tileSize},
			},
		)

		// Кэшируем изображение
		repo.imagesCache[imageID] = spriteImage
	}

	// Копируем данные анимаций и конвертируем в старый формат
	for animationID, animationConfig := range config.Animations {
		// Сохраняем полную конфигурацию
		repo.animationsConfig[animationID] = animationConfig

		// Конвертируем новый формат в старый формат AnimationData
		var animationFrames types.AnimationData
		for _, frameID := range animationConfig.Frames {
			animationFrames = append(animationFrames, types.AnimationDataFrame{
				Image:    frameID,
				Duration: animationConfig.Duration,
			})
		}
		repo.animationsData[animationID] = animationFrames
	}

	return repo, nil
}

func (tr *TilesetDataRepository) GetImage(
	id string,
) (image.Image, error) {
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

// GetAnimationConfig получает конфигурацию анимации
func (tr *TilesetDataRepository) GetAnimationConfig(id string) (types.AnimationConfig, error) {
	config, exists := tr.animationsConfig[id]
	if !exists {
		return types.AnimationConfig{}, fmt.Errorf("animation config '%s' not found", id)
	}
	return config, nil
}
