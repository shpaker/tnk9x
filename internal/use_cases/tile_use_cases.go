package use_cases

import (
	"fmt"
	"image"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/types"
)

// TilesUseCases содержит бизнес-логику для работы с тайлами и анимациями
type TilesUseCases struct {
	tilesRepository      processed.ITilesetRepository
	animationsRepo       game.IAnimationsRepository
	spawnerTilesetRepo   processed.ITilesetRepository
	explosionTilesetRepo processed.ITilesetRepository
	tileService          *services.TileService
	animationService     *services.AnimationService
}

// NewTilesUseCases создает новый экземпляр TilesUseCases
func NewTilesUseCases(
	tilesRepository processed.ITilesetRepository,
) *TilesUseCases {
	return &TilesUseCases{
		tilesRepository:  tilesRepository,
		tileService:      services.NewTileService(tilesRepository),
		animationService: services.NewAnimationService(),
	}
}

// NewTilesUseCasesWithAnimations создает новый экземпляр TilesUseCases с поддержкой анимаций
func NewTilesUseCasesWithAnimations(
	tilesRepository processed.ITilesetRepository,
	animationsRepo game.IAnimationsRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	explosionTilesetRepo processed.ITilesetRepository,
) *TilesUseCases {
	tuc := &TilesUseCases{
		tilesRepository:      tilesRepository,
		animationsRepo:       animationsRepo,
		spawnerTilesetRepo:   spawnerTilesetRepo,
		explosionTilesetRepo: explosionTilesetRepo,
		animationService:     services.NewAnimationService(),
	}

	tuc.tileService = services.NewTileServiceWithSpecialRepos(
		tilesRepository,
		spawnerTilesetRepo,
		explosionTilesetRepo,
	)

	return tuc
}

// GetImage возвращает изображение по ID
func (tuc *TilesUseCases) GetImage(id string) (image.Image, error) {
	return tuc.tilesRepository.GetImage(id)
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
			tuc.animationService.UpdateAnimation(animation)
		}
	}
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
	if tuc.spawnerTilesetRepo == nil {
		return nil, fmt.Errorf("spawner tileset repository not initialized")
	}

	tempService := services.NewTileService(tuc.spawnerTilesetRepo)
	animation, err := tempService.CreateAnimationTileFromRepo(
		tuc.spawnerTilesetRepo,
		"spawner",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}

// CreateExplosionAnimation создает анимацию взрыва
func (tuc *TilesUseCases) CreateExplosionAnimation() (*types.TileAnimationEntity, error) {
	if tuc.explosionTilesetRepo == nil {
		return nil, fmt.Errorf("explosion tileset repository not initialized")
	}

	tempService := services.NewTileService(tuc.explosionTilesetRepo)
	animation, err := tempService.CreateAnimationTileFromRepo(
		tuc.explosionTilesetRepo,
		"explosion",
	)
	if err != nil {
		return nil, err
	}

	tuc.AddAnimation(animation)
	return animation, nil
}
