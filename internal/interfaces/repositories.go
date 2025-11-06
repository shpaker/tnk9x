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
	BulletsRepository() IBulletsRepository
	AnimationsRepository() IAnimationsRepository
	TanksRepository() ITanksRepository
}

// IBulletsRepository определяет интерфейс для работы с пулями
type IBulletsRepository interface {
	// AddBullet добавляет пулю в репозиторий
	// Возвращает ошибку если у пули нет owner или если у этого owner уже есть пуля
	AddBullet(bullet types.BulletEntity) error

	// GetAllBullets возвращает все пули
	GetAllBullets() []types.BulletEntity

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
	// === Методы для работы с игроком ===
	SetPlayer(player *types.TankEntity)
	GetPlayer() *types.TankEntity
	HasPlayer() bool

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

// ITilesetRepository определяет интерфейс для работы с тайлсетами
type ITilesetRepository interface {
	// GetImage возвращает изображение по ID из тайлсета
	GetImage(id string) (image.Image, error)
	// GetAnimationData возвращает данные анимации по ID
	GetAnimationData(id string) (types.AnimationData, error)
	// GetAnimationConfig возвращает конфигурацию анимации по ID
	GetAnimationConfig(id string) (types.AnimationConfig, error)
}

// IMapsDataRepository определяет интерфейс для работы с картами уровней
type IMapsDataRepository interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int, tileBaseSize int) (*types.MapEntity, error)
	// GetLevelsCount возвращает количество доступных карт (файлы вида *.bcmap)
	GetLevelsCount() (int, error)
}

// ITilesetRepositoryRegistry определяет интерфейс для реестра тайлсетов
type ITilesetRepositoryRegistry interface {
	// Blocks возвращает репозиторий тайлсетов для блоков
	Blocks() ITilesetRepository
	// Player возвращает репозиторий тайлсетов для игрока
	Player() ITilesetRepository
	// Bullet возвращает репозиторий тайлсетов для пуль
	Bullet() ITilesetRepository
	// Spawner возвращает репозиторий тайлсетов для спавна
	Spawner() ITilesetRepository
	// Explosion возвращает репозиторий тайлсетов для взрыва
	Explosion() ITilesetRepository
	// HQ возвращает репозиторий тайлсетов для базы
	HQ() ITilesetRepository
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
