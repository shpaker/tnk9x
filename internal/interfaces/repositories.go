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
	GetEffectsRepository() IEffectsRepository
}

// IEffectsRepository хранит короткоживущие визуальные эффекты
type IEffectsRepository interface {
	AddEffect(effect *types.EffectEntity)
	GetAllEffects() []*types.EffectEntity
	RemoveEffect(effect *types.EffectEntity)
}

// ISoundEventsRepository хранит очередь звуковых событий кадра
type ISoundEventsRepository interface {
	Add(event types.SoundEntity)
	Drain() []types.SoundEntity
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
	RemoveEnemy(enemy *types.TankEntity)
	GetAllEnemies() []*types.TankEntity

	GetAllTanks() []*types.TankEntity
}

// IWavesRepository — состав вражеских волн по этапам
type IWavesRepository interface {
	GetWave(stage int) (types.StageWave, error)
	GetStages() ([]int, error)
}

type IBonusesRepository interface {
	AddBonus(bonus *types.BonusEntity)
	GetAllBonuses() []*types.BonusEntity
	RemoveBonus(bonus *types.BonusEntity) error
	RemoveBonusesWithoutOwner()
}

type IMapsDataRepository interface {
	GetLevel(num int, tileBaseSize int) (*types.MapEntity, error)

	GetLevelsCount() (int, error)
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
