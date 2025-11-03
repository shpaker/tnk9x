package use_cases

import (
	"fmt"
	"image"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// TilesUseCases содержит бизнес-логику для работы с тайлами и анимациями
type TilesUseCases struct {
	tilesRepository            interfaces.ITilesetRepository
	animationsRepository       interfaces.IAnimationsRepository
	spawnerTilesetRepository   interfaces.ITilesetRepository
	explosionTilesetRepository interfaces.ITilesetRepository
	tileService                interfaces.ITileService
	animationService           interfaces.IAnimationService
}

// NewTilesUseCases создает новый экземпляр TilesUseCases
func NewTilesUseCases(
	tilesRepository interfaces.ITilesetRepository,
	tileService interfaces.ITileService,
	animationService interfaces.IAnimationService,
) *TilesUseCases {
	return &TilesUseCases{
		tilesRepository:  tilesRepository,
		tileService:      tileService,
		animationService: animationService,
	}
}

// NewTilesUseCasesWithAnimations создает новый экземпляр TilesUseCases с поддержкой анимаций
func NewTilesUseCasesWithAnimations(
	tilesRepository interfaces.ITilesetRepository,
	animationsRepository interfaces.IAnimationsRepository,
	spawnerTilesetRepository interfaces.ITilesetRepository,
	explosionTilesetRepository interfaces.ITilesetRepository,
	tileService interfaces.ITileService,
	animationService interfaces.IAnimationService,
) *TilesUseCases {
	tuc := &TilesUseCases{
		tilesRepository:            tilesRepository,
		animationsRepository:       animationsRepository,
		spawnerTilesetRepository:   spawnerTilesetRepository,
		explosionTilesetRepository: explosionTilesetRepository,
		tileService:                tileService,
		animationService:           animationService,
	}

	return tuc
}

// GetImage возвращает изображение по ID
func (tuc *TilesUseCases) GetImage(id string) (image.Image, error) {
	return tuc.tilesRepository.GetImage(id)
}

// CreateStaticTile создает статический тайл по ID изображения
func (tuc *TilesUseCases) CreateStaticTile(
	id string,
) (types.IImageProvider, error) {
	// Проверяем, что изображение существует
	_, err := tuc.tilesRepository.GetImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

// CreateAnimationTile создает анимированный тайл по ID анимации
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

// === Методы для работы с анимациями из AnimationUseCases ===

// AddAnimation добавляет анимацию в репозиторий
func (tuc *TilesUseCases) AddAnimation(
	animation *image_providers.AnimationProvider,
) {
	if tuc.animationsRepository == nil {
		return
	}
	tuc.animationsRepository.AddAnimation(animation)
}

// UpdateAnimations обновляет все анимации в репозитории
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

// StartAnimation запускает анимацию объекта
func (tuc *TilesUseCases) StartAnimation(
	animation *image_providers.AnimationProvider,
) {
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

// StopAnimation останавливает анимацию объекта
func (tuc *TilesUseCases) StopAnimation(
	animation *image_providers.AnimationProvider,
) {
	if animation == nil {
		return
	}
	animation.IsAnimating = false
}

// CreateSpawnAnimation создает анимацию спавна
func (tuc *TilesUseCases) CreateSpawnAnimation() (*image_providers.AnimationProvider, error) {
	if tuc.spawnerTilesetRepository == nil {
		return nil, fmt.Errorf("spawner tileset repository not initialized")
	}

	// Используем tileService для создания анимации из специального репозитория
	animation, err := tuc.tileService.CreateAnimationTileFromRepo(
		tuc.spawnerTilesetRepository,
		"spawner",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}

// CreateExplosionAnimation создает анимацию взрыва
func (tuc *TilesUseCases) CreateExplosionAnimation() (*image_providers.AnimationProvider, error) {
	if tuc.explosionTilesetRepository == nil {
		return nil, fmt.Errorf("explosion tileset repository not initialized")
	}

	// Используем tileService для создания анимации из специального репозитория
	animation, err := tuc.tileService.CreateAnimationTileFromRepo(
		tuc.explosionTilesetRepository,
		"explosion",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}
