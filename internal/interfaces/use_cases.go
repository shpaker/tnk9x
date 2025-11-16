package interfaces

import (
	"image"

	"golang.org/x/image/font/opentype"

	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

type IBulletUseCases interface {
	ShootBullet(tank *types.TankEntity) error
	UpdateBullets(dt float64) error
	GetBullets() []*types.BulletEntity
	RemoveBullet(index int) error
}

type IMapUseCases interface {
	GetBlocks() types.MapBlocks
	RemoveBlock(block *types.BlockEntity) error
	GetSizePx() types.Size
	SpawnBonus(baseSizePx uint) (*types.BonusEntity, error)
	GetRandomBonusSpawnPosition() types.Position
}

type ICollisionUseCases interface {
	UpdateCollisions()
	HasTankCollision(candidate *types.TankEntity) bool
	IsSpawnerBlocked(position types.Position, size types.Size) bool
}

type ITilesUseCases interface {
	CreateStaticTile(id string) (IImageProvider, error)
	CreateAnimationTile(id string) (*image_providers.AnimationProvider, error)
	CreateSpawnAnimation() (*image_providers.AnimationProvider, error)
	CreateExplosionAnimation() (*image_providers.AnimationProvider, error)
	GetImage(id string) (image.Image, error)
	AddAnimation(animation *image_providers.AnimationProvider)
	UpdateAnimations()
	StartAnimation(animation *image_providers.AnimationProvider)
}

type ITankCommonUseCases interface {
	Update(tank *types.TankEntity, dt float64) error
	UpdateAllTanks(dt float64) error
	GetAllTanks() []*types.TankEntity
	GetAllPlayerTanks() []*types.TankEntity
	IsAnyPlayerTankMoving() bool
	LevelUp(tank *types.TankEntity)
	LevelDown(tank *types.TankEntity)
	GetTankAnimationName(tank *types.TankEntity) string
}

type ITankAnimationNameProvider interface {
	GetTankAnimationName(tank *types.TankEntity) string
}

type IRenderUseCases interface {
	IsTankSpawnAnimationFinished(tank *types.TankEntity) bool
	IsTankExplosionAnimationFinished(tank *types.TankEntity) bool
	UpdateTankAnimation(tank *types.TankEntity)
	SyncTankAnimationWithState(tank *types.TankEntity)
	UpdateBlink(blinkObjects []types.IBlink)
}

type ITankLifecycleUseCases interface {
	OnStageSetUpEnemiesSpawn() ([3]*types.TankEntity, error)
	SpawnEnemy(index *int, ignoreRespawnDelay bool) (*types.TankEntity, error)
	SpawnEnemyWithLevel(
		index *int,
		ignoreRespawnDelay bool,
		remainingEnemies uint,
	) (*types.TankEntity, error)
	SpawnPlayer1() (*types.TankEntity, error)
	GetPlayerTank(num types.PlayerTankNum) *types.TankEntity
	SetPlayerTank(num types.PlayerTankNum, tank *types.TankEntity)
	SpawnPlayer2() (*types.TankEntity, error)
	Explode(tank *types.TankEntity) error
	IsSpawnFinished(tank *types.TankEntity, currentTime float64)
	IsExplosionFinished(tank *types.TankEntity)
	UpdateAllTanksLifecycle() error
}

type ITankActionsUseCases interface {
	Update(tank *types.TankEntity, dt float64) error
	Rotate(tank *types.TankEntity, direction types.Direction) error
	Move(tank *types.TankEntity) error
	Stop(tank *types.TankEntity, byCollision bool)
	Shoot(tank *types.TankEntity) error
	ApplyDecision(tank *types.TankEntity, decision types.EnemyAIDecision)
	SetMinXPosition(tank *types.TankEntity)
	SetMaxXPosition(tank *types.TankEntity)
	SetMinYPosition(tank *types.TankEntity)
	SetMaxYPosition(tank *types.TankEntity)
}

type IAIUseCases interface {
	ExecuteAI(tank *types.TankEntity) (types.EnemyAIDecision, error)
	Close()
}

type IHQUseCases interface {
	GetHQ() *types.HQEntity
	Explode(hq *types.HQEntity) error
	IsExplosionFinished(hq *types.HQEntity)
	IsDestroyed() bool
}

type IFontUseCases interface {
	GetFont() (*opentype.Font, error)
}

type IStageUseCases interface {
	SpawnPlayerTank(role types.TankRole) *types.TankEntity
	SpawnInitialEnemyTanks() []*types.TankEntity
	TrySpawnEnemy() *types.TankEntity
	TryRespawnPlayersTanks() (*types.TankEntity, *types.TankEntity)
	GetPlayersTanks() []*types.TankEntity
	UpdateGameObjects(dt float64)
	TogglePause()
	IsPaused() bool
	PauseStageState()
	ResumeStageState()
	IsStageWon() bool
	IsStageLost() bool
	IsStageFinished() bool
}

type ISpecsUseCases interface {
	GetTankSpecs(isEnemy bool, level uint) *types.SpecsEntity
	GetEnemyLevelByRemainingCount(remainingEnemies uint) uint
}
