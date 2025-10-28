package use_cases

import (
	"fmt"
	"image"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
)

// TilesUseCases содержит бизнес-логику для работы с тайлами и анимациями
type TilesUseCases struct {
	tilesRepository      processed.ITilesetRepository
	animationsRepo       game.IAnimationsRepository
	spawnerTilesetRepo   processed.ITilesetRepository
	explosionTilesetRepo processed.ITilesetRepository
}

// NewTilesUseCases создает новый экземпляр TilesUseCases
func NewTilesUseCases(
	tilesRepository processed.ITilesetRepository,
) *TilesUseCases {
	return &TilesUseCases{
		tilesRepository: tilesRepository,
	}
}

// NewTilesUseCasesWithAnimations создает новый экземпляр TilesUseCases с поддержкой анимаций
func NewTilesUseCasesWithAnimations(
	tilesRepository processed.ITilesetRepository,
	animationsRepo game.IAnimationsRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	explosionTilesetRepo processed.ITilesetRepository,
) *TilesUseCases {
	return &TilesUseCases{
		tilesRepository:      tilesRepository,
		animationsRepo:       animationsRepo,
		spawnerTilesetRepo:   spawnerTilesetRepo,
		explosionTilesetRepo: explosionTilesetRepo,
	}
}

// GetImage возвращает изображение по ID
func (tuc *TilesUseCases) GetImage(id string) (image.Image, error) {
	return tuc.tilesRepository.GetImage(id)
}

// GetTileAnimationFrames возвращает данные анимации по ID
func (tuc *TilesUseCases) GetTileAnimationFrames(id string) (types.AnimationData, error) {
	return tuc.tilesRepository.GetAnimationData(id)
}

// CreateStaticTile создает статический тайл по ID изображения
func (tuc *TilesUseCases) CreateStaticTile(
	id string,
) (types.IImageIdGetter, error) {
	// Проверяем, что изображение существует
	_, err := tuc.tilesRepository.GetImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &types.TileStaticEntity{
		ImageId: id,
	}, nil
}

// CreateAnimationTile создает анимированный тайл по ID анимации
func (tuc *TilesUseCases) CreateAnimationTile(id string) (*types.TileAnimationEntity, error) {
	// Получаем конфигурацию анимации
	config, err := tuc.tilesRepository.GetAnimationConfig(id)
	if err != nil {
		return nil, fmt.Errorf("animation config '%s' not found: %w", id, err)
	}

	// Получаем данные анимации
	animationFrames, err := tuc.tilesRepository.GetAnimationData(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	// Создаем анимацию с учетом конфигурации
	var animation *types.TileAnimationEntity

	// Проверяем, есть ли offset в конфиге
	hasOffset := config.Offset[0] != 0 || config.Offset[1] != 0

	if config.Repeats == nil {
		// Бесконечная анимация
		if hasOffset {
			animation = types.NewTileAnimationEntityWithOffset(animationFrames, config.Offset)
		} else {
			animation = types.NewTileAnimationEntity(animationFrames)
		}
	} else {
		// Анимация с ограниченным количеством повторений
		if hasOffset {
			animation = types.NewTileAnimationEntityWithLoopsAndOffset(animationFrames, *config.Repeats, config.Offset)
		} else {
			animation = types.NewTileAnimationEntityWithLoops(animationFrames, *config.Repeats)
		}
	}

	return animation, nil
}

// === Методы для работы с анимациями из AnimationUseCases ===

// AddAnimation добавляет анимацию в репозиторий
func (tuc *TilesUseCases) AddAnimation(animation *types.TileAnimationEntity) {
	if tuc.animationsRepo == nil {
		return
	}
	tuc.animationsRepo.AddAnimation(animation)
}

// UpdateAnimations обновляет все анимации в репозитории
func (tuc *TilesUseCases) UpdateAnimations() {
	if tuc.animationsRepo == nil {
		return
	}
	animations := tuc.animationsRepo.GetAllAnimations()
	for _, animation := range animations {
		if animation != nil {
			animation.UpdateAnimation()
		}
	}
}

// StartAnimation запускает анимацию объекта
func (tuc *TilesUseCases) StartAnimation(animation *types.TileAnimationEntity) {
	if animation == nil {
		return
	}
	animation.StartAnimation()
}

// StopAnimation останавливает анимацию объекта
func (tuc *TilesUseCases) StopAnimation(animation *types.TileAnimationEntity) {
	if animation == nil {
		return
	}
	animation.StopAnimation()
}

// CreateSpawnAnimation создает анимацию спавна
func (tuc *TilesUseCases) CreateSpawnAnimation() (*types.TileAnimationEntity, error) {
	if tuc.spawnerTilesetRepo == nil {
		return nil, fmt.Errorf("spawner tileset repository not initialized")
	}
	tilesUseCases := NewTilesUseCases(tuc.spawnerTilesetRepo)
	spawnAnimation, err := tilesUseCases.CreateAnimationTile("spawner")
	if err != nil {
		return nil, err
	}
	tuc.AddAnimation(spawnAnimation)
	return spawnAnimation, nil
}

// CreateExplosionAnimation создает анимацию взрыва
func (tuc *TilesUseCases) CreateExplosionAnimation() (*types.TileAnimationEntity, error) {
	if tuc.explosionTilesetRepo == nil {
		return nil, fmt.Errorf("explosion tileset repository not initialized")
	}
	tilesUseCases := NewTilesUseCases(tuc.explosionTilesetRepo)
	explosionAnimation, err := tilesUseCases.CreateAnimationTile("explosion")
	if err != nil {
		return nil, err
	}
	tuc.AddAnimation(explosionAnimation)
	return explosionAnimation, nil
}
