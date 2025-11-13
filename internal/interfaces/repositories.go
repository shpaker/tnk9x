package interfaces

import (
	"image"

	"github.com/shpaker/tnk25/internal/types"
	image_providers "github.com/shpaker/tnk25/internal/types/image_providers"
)

type IGameRepositoriesRegistry interface {
	GetBulletsRepository() IBulletsRepository
	GetAnimationsRepository() IAnimationsRepository
	GetTanksRepository() ITanksRepository
	GetBonusesRepository() IBonusesRepository
}

type IBulletsRepository interface {
	AddBullet(bullet *types.BulletEntity) error

	GetAllBullets() []*types.BulletEntity

	RemoveBullet(index int) error
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

	AddEnemy(enemy *types.TankEntity)
	GetAllEnemies() []*types.TankEntity

	GetAllTanks() []*types.TankEntity

	AddTank(tank *types.TankEntity)
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

type ITilesetRepositoryRegistry interface {
	GetBlocksImage(id string) (types.IImageProvider, error)
	GetBlocksAnimationData(id string) (types.AnimationData, error)
	GetBlocksAnimationConfig(id string) (types.AnimationConfig, error)

	GetPlayerImage(id string) (types.IImageProvider, error)
	GetPlayerAnimationData(id string) (types.AnimationData, error)
	GetPlayerAnimationConfig(id string) (types.AnimationConfig, error)

	GetEnemyImage(id string) (types.IImageProvider, error)
	GetEnemyAnimationData(id string) (types.AnimationData, error)
	GetEnemyAnimationConfig(id string) (types.AnimationConfig, error)

	GetBulletImage(id string) (types.IImageProvider, error)
	GetBulletAnimationData(id string) (types.AnimationData, error)
	GetBulletAnimationConfig(id string) (types.AnimationConfig, error)

	GetSpawnerImage(id string) (types.IImageProvider, error)
	GetSpawnerAnimationData(id string) (types.AnimationData, error)
	GetSpawnerAnimationConfig(id string) (types.AnimationConfig, error)

	GetExplosionImage(id string) (types.IImageProvider, error)
	GetExplosionAnimationData(id string) (types.AnimationData, error)
	GetExplosionAnimationConfig(id string) (types.AnimationConfig, error)

	GetHQImage(id string) (types.IImageProvider, error)
	GetHQAnimationData(id string) (types.AnimationData, error)
	GetHQAnimationConfig(id string) (types.AnimationConfig, error)

	GetBonusesImage(id string) (types.IImageProvider, error)
	GetBonusesAnimationData(id string) (types.AnimationData, error)
	GetBonusesAnimationConfig(id string) (types.AnimationConfig, error)

	GetImageData(tilesetType string, id string) (image.Image, error)
}

type IScriptsRepository interface {
	GetScript(name string) (string, error)
}

type IFontsRepository interface {
	GetFont(name string) ([]byte, error)
}

type IFileRepository interface {
	ReadFile(name string) ([]byte, error)

	ReadImage(name string) (image.Image, error)

	CountFiles(dirPath string, pattern string) (int, error)
}
