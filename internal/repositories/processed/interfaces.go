package processed

import (
	"image"

	"github.com/shpaker/gonflict/internal/types"
)

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
	GetLevel(num int) ([]types.BlockEntity, error)
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
}

// IScriptsRepository определяет интерфейс для работы с Lua скриптами
type IScriptsRepository interface {
	// GetScript возвращает скрипт по имени
	GetScript(name string) (string, error)
}
