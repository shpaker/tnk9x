package use_cases

import (
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.ITilesUseCases = (*TilesUseCases)(nil)

type TilesUseCases struct {
	tilesetRegistry      interfaces.ITilesetRepositoryRegistry
	tilesetType          types.TilesetType
	animationsRepository interfaces.IAnimationsRepository
	tileService          interfaces.ITileService
	animationService     interfaces.IAnimationService
}

func NewTilesUseCases(
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	tilesetType types.TilesetType,
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
	tilesetType types.TilesetType,
	animationsRepository interfaces.IAnimationsRepository,
	tileService interfaces.ITileService,
	animationService interfaces.IAnimationService,
) *TilesUseCases {
	return &TilesUseCases{
		tilesetRegistry:      tilesetRegistry,
		tilesetType:          tilesetType,
		animationsRepository: animationsRepository,
		tileService:          tileService,
		animationService:     animationService,
	}
}

func (tuc *TilesUseCases) CreateStaticTile(
	id string,
) (types.IImageProvider, error) {
	_, err := tuc.tilesetRegistry.GetImageData(tuc.tilesetType, id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tuc *TilesUseCases) CreateTankAnimationTile(
	id string,
	isEnemy bool,
) (*image_providers.AnimationProvider, error) {
	return tuc.tileService.CreateAnimationTileFromTileset(
		types.TankTilesetType(isEnemy),
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
	if tuc.animationsRepository == nil {
		return nil, fmt.Errorf("animations repository is not configured")
	}

	animation, err := tuc.tileService.CreateAnimationTileFromTileset(
		types.TilesetTypeSpawner,
		"spawner",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}

func (tuc *TilesUseCases) CreateExplosionAnimation() (*image_providers.AnimationProvider, error) {
	if tuc.animationsRepository == nil {
		return nil, fmt.Errorf("animations repository is not configured")
	}

	animation, err := tuc.tileService.CreateAnimationTileFromTileset(
		types.TilesetTypeExplosion,
		"explosion_tank",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}
