package interfaces

import (
	"image"

	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// ============================================================================
// Game Repositories Interfaces
// ============================================================================

// IGameRepositoriesRegistry определяет интерфейс для реестра игровых репозиториев
type IGameRepositoriesRegistry interface {
	GetBulletsRepository() IBulletsRepository
	GetAnimationsRepository() IAnimationsRepository
	GetTanksRepository() ITanksRepository
}

// IBulletsRepository определяет интерфейс для работы с пулями
type IBulletsRepository interface {
	// AddBullet добавляет пулю в репозиторий
	// Возвращает ошибку если у пули нет owner или если у этого owner уже есть пуля
	AddBullet(bullet *types.BulletEntity) error

	// GetAllBullets возвращает все пули
	GetAllBullets() []*types.BulletEntity

	// RemoveBullet удаляет пулю по индексу
	RemoveBullet(index int) error
}

// IAnimationsRepository определяет интерфейс для работы с анимациями
type IAnimationsRepository interface {
	// AddAnimation добавляет анимацию в репозиторий
	AddAnimation(animation *image_providers.AnimationProvider)

	// GetAllAnimations возвращает все анимации
	GetAllAnimations() []*image_providers.AnimationProvider
}

// ITanksRepository определяет интерфейс для работы с танками
type ITanksRepository interface {
	// === Методы для работы с игроками по номеру ===
	SetPlayer(num types.PlayerTankNum, player *types.TankEntity)
	GetPlayer(num types.PlayerTankNum) *types.TankEntity
	HasPlayer(num types.PlayerTankNum) bool
	GetAllPlayers() []*types.TankEntity

	// === Методы для работы с врагами ===
	AddEnemy(enemy *types.TankEntity)
	GetAllEnemies() []*types.TankEntity

	// === Методы для работы со всеми танками ===
	GetAllTanks() []*types.TankEntity

	// === Методы для обратной совместимости ===
	AddTank(tank *types.TankEntity)
}

// ============================================================================
// Processed Repositories Interfaces
// ============================================================================

// IMapsDataRepository определяет интерфейс для работы с картами уровней
type IMapsDataRepository interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int, tileBaseSize int) (*types.MapEntity, error)
	// GetLevelsCount возвращает количество доступных карт (файлы вида *.bcmap)
	GetLevelsCount() (int, error)
}

// ITilesetRepositoryRegistry определяет интерфейс для фасада тайлсетов
type ITilesetRepositoryRegistry interface {
	// === Методы для блоков ===
	GetBlocksImage(id string) (types.IImageProvider, error)
	GetBlocksAnimationData(id string) (types.AnimationData, error)
	GetBlocksAnimationConfig(id string) (types.AnimationConfig, error)

	// === Методы для игрока ===
	GetPlayerImage(id string) (types.IImageProvider, error)
	GetPlayerAnimationData(id string) (types.AnimationData, error)
	GetPlayerAnimationConfig(id string) (types.AnimationConfig, error)

	// === Методы для врагов ===
	GetEnemyImage(id string) (types.IImageProvider, error)
	GetEnemyAnimationData(id string) (types.AnimationData, error)
	GetEnemyAnimationConfig(id string) (types.AnimationConfig, error)

	// === Методы для пуль ===
	GetBulletImage(id string) (types.IImageProvider, error)
	GetBulletAnimationData(id string) (types.AnimationData, error)
	GetBulletAnimationConfig(id string) (types.AnimationConfig, error)

	// === Методы для спавна ===
	GetSpawnerImage(id string) (types.IImageProvider, error)
	GetSpawnerAnimationData(id string) (types.AnimationData, error)
	GetSpawnerAnimationConfig(id string) (types.AnimationConfig, error)

	// === Методы для взрыва ===
	GetExplosionImage(id string) (types.IImageProvider, error)
	GetExplosionAnimationData(id string) (types.AnimationData, error)
	GetExplosionAnimationConfig(id string) (types.AnimationConfig, error)

	// === Методы для базы ===
	GetHQImage(id string) (types.IImageProvider, error)
	GetHQAnimationData(id string) (types.AnimationData, error)
	GetHQAnimationConfig(id string) (types.AnimationConfig, error)

	// GetImageData возвращает image.Image по типу тайлсета и ID (для внутреннего использования)
	// tilesetType передается как string для избежания циклических зависимостей
	GetImageData(tilesetType string, id string) (image.Image, error)
}

// IScriptsRepository определяет интерфейс для работы с Lua скриптами
type IScriptsRepository interface {
	// GetScript возвращает скрипт по имени
	GetScript(name string) (string, error)
}

// IFontsRepository определяет интерфейс для работы со шрифтами
type IFontsRepository interface {
	// GetFont возвращает данные шрифта по имени (без расширения .ttf)
	GetFont(name string) ([]byte, error)
}

// ============================================================================
// Raw Repositories Interfaces
// ============================================================================

// IFileRepository определяет интерфейс для работы с файлами
type IFileRepository interface {
	// ReadFile читает файл и возвращает его содержимое в виде байтов
	ReadFile(name string) ([]byte, error)

	// ReadImage читает изображение (добавляет расширение .png автоматически)
	ReadImage(name string) (image.Image, error)

	// CountFiles возвращает количество файлов в указанной директории по маске (например, "*.bcmap")
	CountFiles(dirPath string, pattern string) (int, error)
}
