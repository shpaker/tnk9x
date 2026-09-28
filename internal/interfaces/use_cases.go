package interfaces

import (
	"image"
	"image/color"

	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

type IBulletUseCases interface {
	// ShootBullet выпускает пулю танка; fired — false, если танк
	// неактивен или его пули ещё в полёте (лимит — не ошибка)
	ShootBullet(tank *types.TankEntity) (fired bool, err error)
	UpdateBullets(dt float64) error
	GetBullets() []*types.BulletEntity
	RemoveBullet(bullet *types.BulletEntity) error
}

type IMapUseCases interface {
	GetBlocks() types.MapBlocks
	RemoveBlock(block *types.BlockEntity) error
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
	CreateTankAnimationTile(
		id string,
		isEnemy bool,
	) (*image_providers.AnimationProvider, error)
	AddAnimation(animation *image_providers.AnimationProvider)
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
	// SpawnEnemy создаёт врага заданного уровня на спаунере с индексом
	// spawnerIndex; занятость спаунера проверяет вызывающий
	SpawnEnemy(spawnerIndex int, level uint) (*types.TankEntity, error)
	SpawnPlayer1() (*types.TankEntity, error)
	GetPlayerTank(num types.PlayerTankNum) *types.TankEntity
	SetPlayerTank(num types.PlayerTankNum, tank *types.TankEntity)
	SpawnPlayer2() (*types.TankEntity, error)
	Explode(tank *types.TankEntity) error
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

// IAIUseCases собирает снимок мира для вражеского танка и получает
// решение от AI-скрипта
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
	// GetStageResult — итог уровня для подсчёта звёзд
	GetStageResult() types.StageResult
}

// IWaveUseCases — выдача врагов по волнам сценария уровня
type IWaveUseCases interface {
	// NextTank — следующий танк сценария, если его волна уже началась;
	// false — врагов не осталось или следующая волна ещё ждёт условия
	NextTank(session *session_entities.StageSessionEntity) (
		types.WaveTank,
		bool,
	)
	// CommitSpawn сдвигает курсор волн после спауна танка из NextTank
	// и запускает паузу волны до следующего спауна
	CommitSpawn(
		session *session_entities.StageSessionEntity,
		spawnerIndex int,
	)
}

// IEnemySpawnSelectionUseCases — выбор точки спауна врага
type IEnemySpawnSelectionUseCases interface {
	// SelectSpawner — индекс свободного спаунера; false — все заняты
	SelectSpawner(session *session_entities.StageSessionEntity) (int, bool)
	// GetSpawnersCount — число спаунеров врагов
	GetSpawnersCount() int
}

// IProgressionUseCases — звёзды за уровни и открытие кампании
type IProgressionUseCases interface {
	GetCampaign() *types.CampaignEntity
	// CalcStars — звёзды за итог уровня: победа, без потерь, на время
	CalcStars(result types.StageResult, level *types.LevelEntity) uint
	// RecordResult сохраняет звёзды, если результат лучше прежнего
	RecordResult(level int, stars uint) error
	GetLevelStars(level int) uint
	IsLevelUnlocked(level int) bool
	GetPackStatus(packIndex int) types.PackStatus
	// NextLevel — следующий уровень кампании и открыт ли он
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

// ILightingUseCases — источники света кадра и материалы поверхностей
// для графических эффектов
type ILightingUseCases interface {
	// GetLights возвращает не больше types.MaxLights источников
	// в координатах поля, по убыванию приоритета
	GetLights() []types.LightEntity
	// GetMaterial возвращает свойства поверхности блока для освещения
	GetMaterial(blockType types.BlockType) types.SurfaceMaterial
	// GetViewers возвращает зрителей — танки игроков — в координатах поля
	GetViewers() []types.ViewerEntity
	// UpdateHeadlights доворачивает фары танков игроков за стволом;
	// вызывается раз в тик
	UpdateHeadlights()
}

// IVisionUseCases — зрение игроков: насколько видны враги и пули
// в зависимости от фары, периферии, дистанции и зданий на линии взгляда
type IVisionUseCases interface {
	// UpdateVisibility пересчитывает видимость врагов и всех пуль;
	// вызывается раз в тик
	UpdateVisibility()
}

// IVisualEffectsUseCases — графические эффекты игровых событий:
// частицы, вспышки, тряска экрана и отдача выстрела
type IVisualEffectsUseCases interface {
	// RequestEffect ставит событие в очередь кадра
	RequestEffect(event types.VisualEventEntity)
	// Update разбирает события кадра и продвигает эффекты на тик
	Update()
	// GetParticles возвращает живые частицы в координатах поля
	GetParticles() []types.ParticleEntity
	// GetFlashLights возвращает источники света вспышек в координатах поля
	GetFlashLights() []types.LightEntity
	// GetShakeOffset возвращает смещение экрана от тряски в целых пикселях
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

// ILevelSelectUseCases — навигация по кампании на экране выбора уровня
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
