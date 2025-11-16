package processed

import (
	"fmt"
	"image"

	"gopkg.in/yaml.v3"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

type TilesetDataRepository struct {
	fileRepo         interfaces.IFileRepository
	imagesCache      map[string]image.Image
	animationsData   map[string]types.AnimationData
	animationsConfig map[string]types.AnimationConfig
}

func NewTilesetDataRepository(
	fileRepo interfaces.IFileRepository,
	tilesetName string,
) (*TilesetDataRepository, error) {
	configData, err := fileRepo.ReadFile(tilesetName + ".yml")
	if err != nil {
		return nil, err
	}

	var config types.TilesetDataConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	img, err := fileRepo.ReadImage(tilesetName)
	if err != nil {
		return nil, err
	}

	repo := &TilesetDataRepository{
		fileRepo:         fileRepo,
		imagesCache:      make(map[string]image.Image),
		animationsData:   make(map[string]types.AnimationData),
		animationsConfig: make(map[string]types.AnimationConfig),
	}

	tileSize := config.Size
	for imageID, coords := range config.Images {

		spriteImage := img.(interfaces.ISubImageProvider).SubImage(
			image.Rectangle{
				Min: image.Point{
					X: coords[0] * tileSize,
					Y: coords[1] * tileSize,
				},
				Max: image.Point{
					X: (coords[0] + 1) * tileSize,
					Y: (coords[1] + 1) * tileSize,
				},
			},
		)

		repo.imagesCache[imageID] = spriteImage
	}

	for animationID, animationConfig := range config.Animations {

		repo.animationsConfig[animationID] = animationConfig

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

func (tr *TilesetDataRepository) getImage(
	id string,
) (image.Image, error) {
	if cachedImage, exists := tr.imagesCache[id]; exists {
		return cachedImage, nil
	}

	return nil, fmt.Errorf("image '%s' not found", id)
}

func (tr *TilesetDataRepository) getAnimationData(
	id string,
) (types.AnimationData, error) {
	animationData, exists := tr.animationsData[id]
	if !exists {
		return nil, fmt.Errorf("animation '%s' not found", id)
	}

	return animationData, nil
}

func (tr *TilesetDataRepository) getAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	config, exists := tr.animationsConfig[id]
	if !exists {
		return types.AnimationConfig{}, fmt.Errorf(
			"animation config '%s' not found",
			id,
		)
	}
	return config, nil
}
