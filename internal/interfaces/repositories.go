package interfaces

import (
	"image"

	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

type IGameRepositoriesRegistry interface {
	GetBulletsRepository() IBulletsRepository
	GetAnimationsRepository() IAnimationsRepository
	GetTanksRepository() ITanksRepository
	GetBonusesRepository() IBonusesRepository
	GetSoundEventsRepository() ISoundEventsRepository
	GetVisualEffectsRepository() IVisualEffectsRepository
}

// ISoundEventsRepository хранит очередь звуковых событий кадра
type ISoundEventsRepository interface {
	Add(event types.SoundEntity)
	Drain() []types.SoundEntity
}

// IVisualEffectsRepository хранит runtime-состояние графических эффектов
// уровня: очередь визуальных событий кадра, частицы, вспышки и тряску
type IVisualEffectsRepository interface {
	AddEvent(event types.VisualEventEntity)
	// DrainEvents возвращает накопленные события и очищает очередь
	DrainEvents() []types.VisualEventEntity

	// AddParticle добавляет частицу; сверх лимита частица отбрасывается
	AddParticle(particle types.ParticleEntity)
	// GetParticles возвращает частицы для изменения на месте
	GetParticles() []types.ParticleEntity
	// SetParticles заменяет набор частиц, например уплотнённым срезом
	SetParticles(particles []types.ParticleEntity)

	AddFlash(flash types.FlashEntity)
	GetFlashes() []types.FlashEntity
	SetFlashes(flashes []types.FlashEntity)

	GetScreenShake() *types.ScreenShakeEntity
}

type IBulletsRepository interface {
	AddBullet(bullet *types.BulletEntity) error

	GetAllBullets() []*types.BulletEntity

	RemoveBullet(bullet *types.BulletEntity) error
}

type IAnimationsRepository interface {
	AddAnimation(animation *image_providers.AnimationProvider)

	GetAllAnimations() []*image_providers.AnimationProvider
}

type ITanksRepository interface {
	SetPlayer(num types.PlayerTankNum, player *types.TankEntity)
	GetPlayer(num types.PlayerTankNum) *types.TankEntity
	HasPlayer(num types.PlayerTankNum) bool
	GetAllPlayers() []*types.TankEntity
	GetActivePlayerTanks() []*types.TankEntity

	AddEnemy(enemy *types.TankEntity)
	GetAllEnemies() []*types.TankEntity

	GetAllTanks() []*types.TankEntity
}

type IBonusesRepository interface {
	AddBonus(bonus *types.BonusEntity)
	GetAllBonuses() []*types.BonusEntity
	RemoveBonus(bonus *types.BonusEntity) error
	RemoveBonusesWithoutOwner()
}

type IMapsDataRepository interface {
	// GetLevel читает уровень; каждый вызов создаёт новую карту
	GetLevel(num int, tileBaseSize int) (*types.LevelEntity, error)
	// HasLevel — существует ли файл уровня
	HasLevel(num int) bool

	GetLevelsCount() (int, error)
}

// ICampaignRepository — кампании: пачки уровней и условия открытия
type ICampaignRepository interface {
	GetCampaign(name string) (*types.CampaignEntity, error)
}

// IStorageRepository — долговременное хранилище пользовательских
// данных (сохранения) по ключу
type IStorageRepository interface {
	// Load возвращает nil без ошибки, если ключа ещё нет
	Load(key string) ([]byte, error)
	Save(key string, data []byte) error
}

// IProgressRepository — прогресс игрока в кампании
type IProgressRepository interface {
	GetProgress() (*types.ProgressEntity, error)
	SaveProgress(progress *types.ProgressEntity) error
}

// ITilesetRepositoryRegistry — единая точка доступа к тайлсетам по типу
type ITilesetRepositoryRegistry interface {
	GetImageData(
		tilesetType types.TilesetType,
		id string,
	) (image.Image, error)
	GetAnimationData(
		tilesetType types.TilesetType,
		id string,
	) (types.AnimationData, error)
	GetAnimationConfig(
		tilesetType types.TilesetType,
		id string,
	) (types.AnimationConfig, error)
	GetImageIDs(tilesetType types.TilesetType) []string
}

type IScriptsRepository interface {
	GetScript(name string) (string, error)
}

// IShadersRepository отдаёт исходники Kage-шейдеров по имени
type IShadersRepository interface {
	GetShader(name string) ([]byte, error)
}

type IFontsRepository interface {
	GetFont(name string) ([]byte, error)
}

type ISoundsRepository interface {
	GetSound(name string) ([]byte, error)
}

type IFileRepository interface {
	ReadFile(name string) ([]byte, error)

	ReadImage(name string) (image.Image, error)

	CountFiles(dirPath string, pattern string) (int, error)
}

type ISubImageProvider interface {
	SubImage(r image.Rectangle) image.Image
}
