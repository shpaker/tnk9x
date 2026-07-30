package interfaces

import (
	"image"
	"image/color"

	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

type IBulletUseCases interface {
	ShootBullet(tank *types.TankEntity) error
	UpdateBullets(dt float64) error
	GetBullets() []*types.BulletEntity
	RemoveBullet(bullet *types.BulletEntity) error
	SpawnImpact(bullet *types.BulletEntity)
	GetImpacts() []*types.EffectEntity
}

type IMapUseCases interface {
	GetBlocks() types.MapBlocks
	RemoveBlock(block *types.BlockEntity) error
	AddBlock(block *types.BlockEntity)
	GetSizePx() types.Size
	GetRandomBonusSpawnPosition() types.Position
	IsIceAt(position types.Position) bool
}

type ICollisionUseCases interface {
	UpdateCollisions()
	IsSpawnerBlocked(position types.Position, size types.Size) bool
}

type ITilesUseCases interface {
	CreateStaticTile(id string) (IImageProvider, error)
	CreateSpawnAnimation() (*image_providers.AnimationProvider, error)
	CreateExplosionAnimation() (*image_providers.AnimationProvider, error)
	CreateBulletExplosionAnimation() (*image_providers.AnimationProvider, error)
	CreateTankAnimationTile(
		id string,
		isEnemy bool,
	) (*image_providers.AnimationProvider, error)
	AddAnimation(animation *image_providers.AnimationProvider)
	RemoveAnimation(animation *image_providers.AnimationProvider)
	UpdateAnimations()
	StartAnimation(animation *image_providers.AnimationProvider)
	StopAnimation(animation *image_providers.AnimationProvider)
}

// ISpriteUseCases — выдача спрайтов по типу тайлсета для рендера
type ISpriteUseCases interface {
	GetImage(
		tilesetType types.TilesetType,
		id string,
	) (image.Image, error)
	GetImageIDs(tilesetType types.TilesetType) []string
}

type ITankCommonUseCases interface {
	Update(tank *types.TankEntity, dt float64) error
	UpdateAllTanks(dt float64) error
	GetAllTanks() []*types.TankEntity
	GetAllPlayerTanks() []*types.TankEntity
	IsAnyPlayerTankMoving() bool
	LevelUp(tank *types.TankEntity)
	LevelDown(tank *types.TankEntity)
}

type IRenderUseCases interface {
	IsTankSpawnAnimationFinished(tank *types.TankEntity) bool
	IsTankExplosionAnimationFinished(tank *types.TankEntity) bool
	UpdateTankAnimation(tank *types.TankEntity)
	SyncTankAnimationWithState(tank *types.TankEntity)
	UpdateBlink(blinkObjects []types.IBlink)
	IsTankVisible(tank *types.TankEntity) bool
	TankHealthOverlay(tank *types.TankEntity) (color.NRGBA, bool)
}

type ITankLifecycleUseCases interface {
	SpawnEnemy(
		spawnIndex uint,
		ignoreBlocked bool,
		level uint,
	) (*types.TankEntity, error)
	SpawnPlayer1(level uint) (*types.TankEntity, error)
	SpawnPlayer2(level uint) (*types.TankEntity, error)
	GetPlayerTank(num types.PlayerTankNum) *types.TankEntity
	SetPlayerTank(num types.PlayerTankNum, tank *types.TankEntity)
	Explode(tank *types.TankEntity) error
	RemoveEnemy(tank *types.TankEntity)
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
	ExecuteAI(
		tank *types.TankEntity,
		context types.EnemyAIContext,
	) (types.EnemyAIDecision, error)
}

type IHQUseCases interface {
	GetHQ() *types.HQEntity
	Explode(hq *types.HQEntity) error
	IsExplosionFinished(hq *types.HQEntity)
	IsDestroyed() bool
}

type IStageUseCases interface {
	SpawnPlayerTank(role types.TankRole) *types.TankEntity
	SpawnInitialEnemyTanks() []*types.TankEntity
	TrySpawnEnemy() *types.TankEntity
	TryRespawnPlayersTanks() (*types.TankEntity, *types.TankEntity)
	GetPlayersTanks() []*types.TankEntity
	UpdateGameObjects(dt float64)
	AreEnemiesFrozen() bool
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
}

type ISoundUseCases interface {
	RequestSound(soundID types.SoundID, loop bool)
	RequestStop(soundID types.SoundID)
	RequestStopAll()
	GetEvents() []types.SoundEntity
}

type IBonusUseCases interface {
	Apply(bonus *types.BonusEntity, tank *types.TankEntity)
	SpawnRandomBonusEntity(position types.Position) *types.BonusEntity
	SpawnBonusOnField()
	VisibleBonuses() []*types.BonusEntity
}

// IFortressUseCases — укрепление кольца вокруг штаба (бонус «лопата»)
type IFortressUseCases interface {
	Apply()
	Update()
}

type IHUDUseCases interface {
	EnemyIconOffsets(
		count uint,
		columns int,
		rows int,
		iconSize int,
	) []types.Position
}

type IStageSelectorUseCases interface {
	Next(selector *types.StageSelectorEntity) uint
	Previous(selector *types.StageSelectorEntity) uint
	String(selector *types.StageSelectorEntity) string
	Select(selector *types.StageSelectorEntity) uint
}
