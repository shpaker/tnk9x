package use_cases

import (
	"fmt"
	"image"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

type TilesUseCases struct {
	tilesetRegistry      interfaces.ITilesetRepositoryRegistry
	tilesetType          processed.TilesetType
	animationsRepository interfaces.IAnimationsRepository
	spawnerTilesetType   processed.TilesetType
	explosionTilesetType processed.TilesetType
	tileService          interfaces.ITileService
	animationService     interfaces.IAnimationService
}

func NewTilesUseCases(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	tilesetType processed.TilesetType,
	tileService interfaces.ITileService,
	animationService interfaces.IAnimationService,
) *TilesUseCases {
	return &TilesUseCases{
		tilesetRegistry:  tilesetRegistry,
		tilesetType:      tilesetType,
		tileService:      tileService,
		animationService: animationService,
	}
}

func NewTilesUseCasesWithAnimations(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	tilesetType processed.TilesetType,
	animationsRepository interfaces.IAnimationsRepository,
	spawnerTilesetType processed.TilesetType,
	explosionTilesetType processed.TilesetType,
	tileService interfaces.ITileService,
	animationService interfaces.IAnimationService,
) *TilesUseCases {
	tuc := &TilesUseCases{
		tilesetRegistry:      tilesetRegistry,
		tilesetType:          tilesetType,
		animationsRepository: animationsRepository,
		spawnerTilesetType:   spawnerTilesetType,
		explosionTilesetType: explosionTilesetType,
		tileService:          tileService,
		animationService:     animationService,
	}

	return tuc
}

func (tuc *TilesUseCases) GetImage(id string) (image.Image, error) {
	return tuc.getImageFromTileset(tuc.tilesetType, id)
}

func (tuc *TilesUseCases) GetTankImage(
	id string,
	isEnemy bool,
) (image.Image, error) {
	tilesetType := processed.TilesetTypePlayer
	if isEnemy {
		tilesetType = processed.TilesetTypeEnemy
	}
	return tuc.getImageFromTileset(tilesetType, id)
}

func (tuc *TilesUseCases) CreateStaticTile(
	id string,
) (types.IImageProvider, error) {
	_, err := tuc.getImageFromTileset(tuc.tilesetType, id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tuc *TilesUseCases) getImageFromTileset(
	tilesetType processed.TilesetType,
	id string,
) (image.Image, error) {
	var provider types.IImageProvider
	var err error

	switch tilesetType {
	case processed.TilesetTypeBlocks:
		provider, err = tuc.tilesetRegistry.GetBlocksImage(id)
	case processed.TilesetTypePlayer:
		provider, err = tuc.tilesetRegistry.GetPlayerImage(id)
	case processed.TilesetTypeEnemy:
		provider, err = tuc.tilesetRegistry.GetEnemyImage(id)
	case processed.TilesetTypeBullet:
		provider, err = tuc.tilesetRegistry.GetBulletImage(id)
	case processed.TilesetTypeSpawner:
		provider, err = tuc.tilesetRegistry.GetSpawnerImage(id)
	case processed.TilesetTypeExplosion:
		provider, err = tuc.tilesetRegistry.GetExplosionImage(id)
	case processed.TilesetTypeHQ:
		provider, err = tuc.tilesetRegistry.GetHQImage(id)
	default:
		return nil, fmt.Errorf("unknown tileset type: %s", tilesetType)
	}

	if err != nil {
		return nil, err
	}

	imageID, err := provider.GetImageID()
	if err != nil {
		return nil, fmt.Errorf("failed to get image ID from provider: %w", err)
	}

	return tuc.tilesetRegistry.GetImageData(string(tilesetType), imageID)
}

func (tuc *TilesUseCases) CreateAnimationTile(
	id string,
) (*image_providers.AnimationProvider, error) {
	config, err := tuc.tileService.GetAnimationConfig(id)
	if err != nil {
		return nil, err
	}

	animationFrames, err := tuc.tileService.GetTileAnimationFrames(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return tuc.tileService.CreateAnimationFromConfig(
		animationFrames,
		config,
	), nil
}

func (tuc *TilesUseCases) CreateTankAnimationTile(
	id string,
	isEnemy bool,
) (*image_providers.AnimationProvider, error) {
	tilesetType := processed.TilesetTypePlayer
	if isEnemy {
		tilesetType = processed.TilesetTypeEnemy
	}
	return tuc.tileService.CreateAnimationTileFromTileset(
		string(tilesetType),
		id,
	)
}

func (tuc *TilesUseCases) AddAnimation(
	animation *image_providers.AnimationProvider,
) {
	if tuc.animationsRepository == nil {
		return
	}
	tuc.animationsRepository.AddAnimation(animation)
}

func (tuc *TilesUseCases) UpdateAnimations() {
	if tuc.animationsRepository == nil {
		return
	}
	animations := tuc.animationsRepository.GetAllAnimations()
	for _, animation := range animations {
		if animation != nil {
			tuc.animationService.UpdateAnimation(animation)
		}
	}
}

func (tuc *TilesUseCases) StartAnimation(
	animation *image_providers.AnimationProvider,
) {
	if animation == nil {
		return
	}
	animation.IsAnimating = true

	_ = animation.LoopCount
}

func (tuc *TilesUseCases) StopAnimation(
	animation *image_providers.AnimationProvider,
) {
	if animation == nil {
		return
	}
	animation.IsAnimating = false
}

func (tuc *TilesUseCases) CreateSpawnAnimation() (*image_providers.AnimationProvider, error) {
	if tuc.spawnerTilesetType == "" {
		return nil, fmt.Errorf("spawner tileset type not initialized")
	}

	animation, err := tuc.tileService.CreateAnimationTileFromTileset(
		string(tuc.spawnerTilesetType),
		"spawner",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}

func (tuc *TilesUseCases) CreateExplosionAnimation() (*image_providers.AnimationProvider, error) {
	if tuc.explosionTilesetType == "" {
		return nil, fmt.Errorf("explosion tileset type not initialized")
	}

	animation, err := tuc.tileService.CreateAnimationTileFromTileset(
		string(tuc.explosionTilesetType),
		"explosion",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}
