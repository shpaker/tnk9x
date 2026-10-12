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
	GetCollisionWorldRepository() ICollisionWorldRepository
}

type ICollisionWorldRepository interface {
	Sync(static []types.IEntityCollider, movers []types.IMover)
	Update(entity types.IEntityCollider)
	Remove(entity types.IEntityCollider)
	Move(
		entity types.IEntityCollider,
		goal types.Position,
		filter types.CollisionFilter,
	) (types.Position, []types.CollisionEntity)
}

type ISoundEventsRepository interface {
	Add(event types.SoundEntity)
	Drain() []types.SoundEntity
}

type IVisualEffectsRepository interface {
	AddEvent(event types.VisualEventEntity)
	DrainEvents() []types.VisualEventEntity

	AddParticle(particle types.ParticleEntity)
	GetParticles() []types.ParticleEntity
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
	GetLevel(num int, tileBaseSize int) (*types.LevelEntity, error)
	GetSceneLevel(name string, tileBaseSize int) (*types.LevelEntity, error)
	HasLevel(num int) bool

	GetLevelsCount() (int, error)
}

type ICampaignRepository interface {
	GetCampaign(name string) (*types.CampaignEntity, error)
}

type IStorageRepository interface {
	// Load returns nil without error if the key does not exist.
	Load(key string) ([]byte, error)
	Save(key string, data []byte) error
}

type IProgressRepository interface {
	GetProgress() (*types.ProgressEntity, error)
	SaveProgress(progress *types.ProgressEntity) error
}

type IInventoryRepository interface {
	GetInventory() (*types.InventoryEntity, error)
	SaveInventory(inventory *types.InventoryEntity) error
}

type ISettingsRepository interface {
	GetSettings() (*types.SettingsEntity, error)
	SaveSettings(settings *types.SettingsEntity) error
}

type IControlsRepository interface {
	GetControls() (*types.ControlsEntity, error)
	SaveControls(controls *types.ControlsEntity) error
}

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

type IShadersRepository interface {
	GetShader(name string) ([]byte, error)
}

type IImagesRepository interface {
	GetImage(name string) (image.Image, error)
}

type ITextsRepository interface {
	GetLocale(language types.Language) ([]byte, error)
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
