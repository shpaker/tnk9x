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

// getTileAnimationFrames возвращает данные анимации по ID
func (tuc *TilesUseCases) getTileAnimationFrames(
	id string,
) (types.AnimationData, error) {
	return tuc.tilesRepository.GetAnimationData(id)
}

// CreateStaticTile создает статический тайл по ID изображения
func (tuc *TilesUseCases) CreateStaticTile(
	id string,
) (types.IImageIDGetter, error) {
	// Проверяем, что изображение существует
	_, err := tuc.tilesRepository.GetImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &types.TileStaticEntity{
		ImageID: id,
	}, nil
}

// CreateAnimationTile создает анимированный тайл по ID анимации
func (tuc *TilesUseCases) CreateAnimationTile(
	id string,
) (*types.TileAnimationEntity, error) {
	config, err := tuc.getAnimationConfig(id)
	if err != nil {
		return nil, err
	}

	animationFrames, err := tuc.getTileAnimationFrames(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return tuc.createAnimationFromConfig(animationFrames, config), nil
}

// getAnimationConfig получает конфигурацию анимации по ID
func (tuc *TilesUseCases) getAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	config, err := tuc.tilesRepository.GetAnimationConfig(id)
	if err != nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"animation config '%s' not found: %w",
			id,
			err,
		)
	}
	return config, nil
}

// createAnimationFromConfig создает анимацию на основе конфигурации и данных кадров
func (tuc *TilesUseCases) createAnimationFromConfig(
	animationFrames types.AnimationData,
	config types.AnimationConfig,
) *types.TileAnimationEntity {
	hasOffset := tuc.hasOffset(config.Offset)
	hasRepeats := config.Repeats != nil

	switch {
	case hasRepeats && hasOffset:
		return types.NewTileAnimationEntityWithLoopsAndOffset(
			animationFrames,
			*config.Repeats,
			config.Offset,
		)
	case hasRepeats:
		return types.NewTileAnimationEntityWithLoops(
			animationFrames,
			*config.Repeats,
		)
	case hasOffset:
		return types.NewTileAnimationEntityWithOffset(
			animationFrames,
			config.Offset,
		)
	default:
		return types.NewTileAnimationEntity(animationFrames)
	}
}

// hasOffset проверяет, есть ли непустое смещение
func (tuc *TilesUseCases) hasOffset(offset [2]float64) bool {
	return offset[0] != 0 || offset[1] != 0
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
			tuc.updateAnimation(animation)
		}
	}
}

// updateAnimation обновляет анимацию на основе тиков
func (tuc *TilesUseCases) updateAnimation(
	animation *types.TileAnimationEntity,
) {
	if animation == nil {
		return
	}

	if len(animation.AnimationFrames) == 0 || !animation.IsAnimating {
		return
	}

	animation.CurrentTick++

	// Проверяем, нужно ли переключить кадр
	if !tuc.shouldAdvanceFrame(animation) {
		return
	}

	nextFrame := tuc.calculateNextFrame(animation)

	// Проверяем завершение циклов и останавливаем анимацию если нужно
	if tuc.checkAndHandleLoopCompletion(animation, nextFrame) {
		return
	}

	animation.CurrentFrame = nextFrame
	animation.CurrentTick = 0
}

// shouldAdvanceFrame проверяет, нужно ли переключать кадр
func (tuc *TilesUseCases) shouldAdvanceFrame(
	animation *types.TileAnimationEntity,
) bool {
	if int(animation.CurrentFrame) >= len(animation.AnimationFrames) {
		return false
	}
	currentFrameDuration := animation.AnimationFrames[animation.CurrentFrame].Duration
	return animation.CurrentTick >= uint(currentFrameDuration)
}

// calculateNextFrame вычисляет следующий кадр анимации
func (tuc *TilesUseCases) calculateNextFrame(
	animation *types.TileAnimationEntity,
) uint {
	return (animation.CurrentFrame + 1) % uint(len(animation.AnimationFrames))
}

// checkAndHandleLoopCompletion проверяет завершение цикла и останавливает анимацию если нужно
// Возвращает true, если анимация была остановлена
func (tuc *TilesUseCases) checkAndHandleLoopCompletion(
	animation *types.TileAnimationEntity,
	nextFrame uint,
) bool {
	if nextFrame == 0 && animation.LoopCount != nil {
		loopsLeft := *animation.LoopCount
		loopsLeft--
		animation.LoopCount = &loopsLeft

		if loopsLeft <= 0 {
			animation.IsAnimating = false
			return true
		}
	}
	return false
}

// StartAnimation запускает анимацию объекта
func (tuc *TilesUseCases) StartAnimation(animation *types.TileAnimationEntity) {
	if animation == nil {
		return
	}
	animation.IsAnimating = true
	// Если у анимации есть repeats, сбрасываем счетчик при каждом запуске
	// Восстанавливаем оригинальное значение repeats из конфигурации
	// Но мы не можем это сделать здесь, так как не храним оригинальное значение
	// Это будет обработано на уровне Use Cases при пересоздании анимации
	_ = animation.LoopCount // Используем для избежания пустой ветки
}

// CreateSpawnAnimation создает анимацию спавна
func (tuc *TilesUseCases) CreateSpawnAnimation() (*types.TileAnimationEntity, error) {
	return tuc.createSpecialAnimation(
		tuc.spawnerTilesetRepo,
		"spawner",
		"spawner tileset repository not initialized",
	)
}

// CreateExplosionAnimation создает анимацию взрыва
func (tuc *TilesUseCases) CreateExplosionAnimation() (*types.TileAnimationEntity, error) {
	return tuc.createSpecialAnimation(
		tuc.explosionTilesetRepo,
		"explosion",
		"explosion tileset repository not initialized",
	)
}

// createSpecialAnimation создает специальную анимацию из указанного репозитория
func (tuc *TilesUseCases) createSpecialAnimation(
	repo processed.ITilesetRepository,
	animationID string,
	errorMsg string,
) (*types.TileAnimationEntity, error) {
	if repo == nil {
		return nil, fmt.Errorf("%s", errorMsg)
	}
	tilesUseCases := NewTilesUseCases(repo)
	animation, err := tilesUseCases.CreateAnimationTile(animationID)
	if err != nil {
		return nil, err
	}
	tuc.AddAnimation(animation)
	return animation, nil
}
