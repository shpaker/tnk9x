package interfaces

import (
	"image"
	"image/color"

	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

type IBulletUseCases interface {
	// fired is false if the tank is inactive or its bullets are still in flight
	// (the limit is not an error).
	ShootBullet(tank *types.TankEntity) (fired bool, err error)
	UpdateBullets(dt float64) error
	GetBullets() []*types.BulletEntity
	RemoveBullet(bullet *types.BulletEntity) error
}

type IMapUseCases interface {
	GetBlocks() types.MapBlocks
	RemoveBlock(block *types.BlockEntity) error
	RestoreBlock(
		block *types.BlockEntity,
		obstacles []types.IEntityCollider,
	) bool
	IsBlockIntact(block *types.BlockEntity) bool
	GetSizePx() types.Size
	GetBorders() []*types.BorderEntity
	GetRandomBonusSpawnPosition() types.Position
	IsIceAt(position types.Position) bool
	IsWaterUnder(position types.Position, size types.Size) bool
}

type ICollisionUseCases interface {
	UpdateCollisions()
	IsSpawnerBlocked(position types.Position, size types.Size) bool
}

type ITilesUseCases interface {
	CreateStaticTile(id string) (IImageProvider, error)
	CreateSpawnAnimation() (*image_providers.AnimationProvider, error)
	CreateExplosionAnimation() (*image_providers.AnimationProvider, error)
	CreateTankAnimationTile(
		id string,
		isEnemy bool,
	) (*image_providers.AnimationProvider, error)
	AddAnimation(animation *image_providers.AnimationProvider)
	UpdateAnimations()
	StartAnimation(animation *image_providers.AnimationProvider)
	StopAnimation(animation *image_providers.AnimationProvider)
}

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
	SetMaxLevel(tank *types.TankEntity)
	IsFrozen(tank *types.TankEntity) bool
}

type IRenderUseCases interface {
	IsTankSpawnAnimationFinished(tank *types.TankEntity) bool
	IsTankExplosionAnimationFinished(tank *types.TankEntity) bool
	UpdateTankAnimation(tank *types.TankEntity)
	SyncTankAnimationWithState(tank *types.TankEntity)
	UpdateBlink(blinkObjects []types.IBlink)
	IsTankBlinking(tank *types.TankEntity) bool
	TankTint(tank *types.TankEntity) (color.NRGBA, bool)
}

type ITankLifecycleUseCases interface {
	SpawnEnemy(spawnerIndex int, level uint) (*types.TankEntity, error)
	SpawnPlayer1(level uint) (*types.TankEntity, error)
	GetPlayerTank(num types.PlayerTankNum) *types.TankEntity
	SetPlayerTank(num types.PlayerTankNum, tank *types.TankEntity)
	SpawnPlayer2(level uint) (*types.TankEntity, error)
	Explode(tank *types.TankEntity) error
	CompleteSpawn(tank *types.TankEntity)
	UpdateAllTanksLifecycle() error
}

type ITankActionsUseCases interface {
	Update(tank *types.TankEntity, dt float64) error
	Rotate(tank *types.TankEntity, direction types.Direction) error
	Move(tank *types.TankEntity) error
	Stop(tank *types.TankEntity, byCollision bool)
	Shoot(tank *types.TankEntity) error
	ApplyDecision(tank *types.TankEntity, decision types.EnemyAIDecision)
}

type IAIUseCases interface {
	ExecuteAI(tank *types.TankEntity, tick int) (types.EnemyAIDecision, error)
}

type IHQUseCases interface {
	GetHQ() *types.HQEntity
	Explode(hq *types.HQEntity) error
	IsExplosionFinished(hq *types.HQEntity)
	IsDestroyed() bool
}

type IStageUseCases interface {
	SpawnPlayerTank(role types.TankRole) *types.TankEntity
	PlacePlayerTank(role types.TankRole) *types.TankEntity
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
	GetStageResult() types.StageResult
	SaveCarryOver()
	BoostCarryOver()
	CanRevivePlayers() bool
	RevivePlayers()
}

type IWaveUseCases interface {
	NextTank(session *session_entities.StageSessionEntity) (
		types.WaveTank,
		bool,
	)
	CommitSpawn(
		session *session_entities.StageSessionEntity,
		spawnerIndex int,
	)
}

type IEnemySpawnSelectionUseCases interface {
	SelectSpawner(session *session_entities.StageSessionEntity) (int, bool)
	GetSpawnersCount() int
}

type IProgressionUseCases interface {
	GetCampaign() *types.CampaignEntity
	CalcStars(result types.StageResult, level *types.LevelEntity) uint
	RecordResult(level int, stars uint) error
	GetLevelStars(level int) uint
	IsLevelUnlocked(level int) bool
	GetPackStatus(packIndex int) types.PackStatus
	NextLevel(level int) (int, bool)
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
	UpdateEffects()
	SpawnRandomBonusEntity(position types.Position) *types.BonusEntity
	VisibleBonuses() []*types.BonusEntity
}

type ILightingUseCases interface {
	GetLights() []types.LightEntity
	GetMaterial(blockType types.BlockType) types.SurfaceMaterial
	UpdateHeadlights()
}

type IVisualEffectsUseCases interface {
	RequestEffect(event types.VisualEventEntity)
	Update()
	GetParticles() []types.ParticleEntity
	GetFlashLights() []types.LightEntity
	GetShakeOffset() types.Position
}

type IHUDUseCases interface {
	EnemyIconOffsets(
		count uint,
		columns int,
		rows int,
		iconSize int,
	) []types.Position
}

type ILevelSelectUseCases interface {
	NewSelector(lastLevel int) *types.LevelSelectorEntity
	MoveLevel(selector *types.LevelSelectorEntity, delta int)
	MovePack(selector *types.LevelSelectorEntity, delta int)
	SetPosition(selector *types.LevelSelectorEntity, position int)
	SelectedLevel(selector *types.LevelSelectorEntity) (int, bool)
	BuildView(
		selector *types.LevelSelectorEntity,
		levels map[int]*types.LevelEntity,
	) types.LevelSelectViewData
}

type ISettingsUseCases interface {
	Items() []types.SettingsItem
	Change(
		settings *types.SettingsEntity,
		item types.SettingsItem,
		step int,
	) error
	SetPlayers(settings *types.SettingsEntity, players uint) error
	MarkHelpShown(settings *types.SettingsEntity) error
	BuildView(
		settings *types.SettingsEntity,
		activeIndex int,
	) types.SettingsViewData
}

type ILocalizationUseCases interface {
	ResolveLanguage(settings *types.SettingsEntity) types.Language
	Apply(settings *types.SettingsEntity) error
}

type IControlsUseCases interface {
	Rows(page types.ControlsPage) []types.ControlsRow
	Bind(
		controls *types.ControlsEntity,
		slot types.ControlsSlot,
		name string,
	) error
	ResetAll(controls *types.ControlsEntity) error
	BuildView(
		controls *types.ControlsEntity,
		cursor types.ControlsCursor,
	) types.ControlsViewData
	HelpRows(controls *types.ControlsEntity) []types.HelpRow
}

type IInventoryUseCases interface {
	HasNoAds() bool
	GetTokens() uint
	// A save error does not give the spent token back.
	SpendToken() (bool, error)
}

type IShopUseCases interface {
	IsAvailable() bool
	GetRows() []types.ShopRow
	IsOnSale(kind types.ProductKind) bool
	RequestPurchase(productID string)
	PollPurchase() (types.PurchaseStatus, error)
	Sync() error
}
