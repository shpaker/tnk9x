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
	// RestoreBlock возвращает блок в исходное целое состояние, если
	// его клетку не занимает ни один из obstacles
	RestoreBlock(
		block *types.BlockEntity,
		obstacles []types.IEntityCollider,
	) bool
	// IsBlockIntact — блок на карте в исходном размере
	IsBlockIntact(block *types.BlockEntity) bool
	GetSizePx() types.Size
	GetRandomBonusSpawnPosition() types.Position
	IsIceAt(position types.Position) bool
	// IsWaterUnder сообщает, перекрывает ли прямоугольник блок воды
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
	// SetMaxLevel сразу даёт танку максимальный уровень (пистолет)
	SetMaxLevel(tank *types.TankEntity)
	// IsFrozen сообщает, заморожена ли сторона танка бонусом-таймером
	IsFrozen(tank *types.TankEntity) bool
}

type IRenderUseCases interface {
	IsTankSpawnAnimationFinished(tank *types.TankEntity) bool
	IsTankExplosionAnimationFinished(tank *types.TankEntity) bool
	UpdateTankAnimation(tank *types.TankEntity)
	SyncTankAnimationWithState(tank *types.TankEntity)
	UpdateBlink(blinkObjects []types.IBlink)
	// IsTankBlinking — мигает ли танк (враг с бонусом или тяжёлый враг)
	IsTankBlinking(tank *types.TankEntity) bool
	// TankTint — тон спрайта мигающего врага в текущем кадре: красный
	// у врага с бонусом, цвет брони у тяжёлого
	TankTint(tank *types.TankEntity) (color.NRGBA, bool)
}

type ITankLifecycleUseCases interface {
	// SpawnEnemy создаёт врага заданного уровня на спаунере с индексом
	// spawnerIndex; занятость спаунера проверяет вызывающий
	SpawnEnemy(spawnerIndex int, level uint) (*types.TankEntity, error)
	// SpawnPlayer1 создаёт танк первого игрока с уровнем прокачки level
	SpawnPlayer1(level uint) (*types.TankEntity, error)
	GetPlayerTank(num types.PlayerTankNum) *types.TankEntity
	SetPlayerTank(num types.PlayerTankNum, tank *types.TankEntity)
	// SpawnPlayer2 создаёт танк второго игрока с уровнем прокачки level
	SpawnPlayer2(level uint) (*types.TankEntity, error)
	Explode(tank *types.TankEntity) error
	// CompleteSpawn завершает появление танка сразу, без анимации
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
	// PlacePlayerTank ставит танк игрока на карту сразу, без анимации
	// появления: уровень начинается с игроком на поле
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
	// GetStageResult — итог уровня для подсчёта звёзд
	GetStageResult() types.StageResult
	// SaveCarryOver запоминает жизни и прокачку танков игроков
	// для переноса на следующий уровень
	SaveCarryOver()
	// BoostCarryOver усиливает снимок переноса: жизнь и уровень
	// танка сверху; без снимка — сверх начальных
	BoostCarryOver()
	// CanRevivePlayers — есть ли второй шанс: уровень проигран
	// потерей всех жизней при целом штабе и оставшихся врагах,
	// второй шанс на этой попытке ещё не использован
	CanRevivePlayers() bool
	// RevivePlayers даёт выбывшим игрокам второй шанс
	RevivePlayers()
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
	// UpdateHeadlights доворачивает фары танков за стволом;
	// вызывается раз в тик
	UpdateHeadlights()
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

// ISettingsUseCases — изменение и сохранение пользовательских настроек
type ISettingsUseCases interface {
	// Items — строки экрана настроек по порядку
	Items() []types.SettingsItem
	// Change меняет значение пункта: переключает флаг или сдвигает
	// громкость на step шагов — и сохраняет настройки
	Change(
		settings *types.SettingsEntity,
		item types.SettingsItem,
		step int,
	) error
	// SetPlayers выбирает режим на одного или двоих игроков;
	// настройки сохраняются, только если режим сменился
	SetPlayers(settings *types.SettingsEntity, players uint) error
	// MarkHelpShown запоминает, что страница «Как играть» показана
	MarkHelpShown(settings *types.SettingsEntity) error
	BuildView(
		settings *types.SettingsEntity,
		activeIndex int,
	) types.SettingsViewData
}

// ILocalizationUseCases — выбор языка интерфейса: язык, выбранный
// в настройках, иначе язык площадки по правилам конфигурации
type ILocalizationUseCases interface {
	// ResolveLanguage — язык интерфейса для настроек
	ResolveLanguage(settings *types.SettingsEntity) types.Language
	// Apply делает активным язык интерфейса для настроек
	Apply(settings *types.SettingsEntity) error
}

// IControlsUseCases — назначение клавиш и кнопок геймпада игроков
// и хоткеев с сохранением раскладки
type IControlsUseCases interface {
	// Rows — строки страницы экрана раскладки по порядку
	Rows(page types.ControlsPage) []types.ControlsRow
	// Bind назначает имя ячейке; занятое другой ячейкой или
	// зарезервированное имя — types.ErrBindingTaken
	Bind(
		controls *types.ControlsEntity,
		slot types.ControlsSlot,
		name string,
	) error
	// ResetAll возвращает всю раскладку к умолчаниям
	ResetAll(controls *types.ControlsEntity) error
	BuildView(
		controls *types.ControlsEntity,
		cursor types.ControlsCursor,
	) types.ControlsViewData
	// HelpRows — таблица управления страницы «Как играть»
	// по текущей раскладке
	HelpRows(controls *types.ControlsEntity) []types.HelpRow
}

// IInventoryUseCases — купленное игроком: отключённая реклама
// и жетоны вместо рекламы за вознаграждение
type IInventoryUseCases interface {
	// HasNoAds — межуровневая реклама отключена покупкой
	HasNoAds() bool
	GetTokens() uint
	// SpendToken списывает жетон и сохраняет инвентарь; false —
	// жетонов нет. Ошибка сохранения жетон не возвращает
	SpendToken() (bool, error)
}

// IShopUseCases — магазин: товары конфигурации, которые есть
// в каталоге площадки, покупка и зачисление купленного
type IShopUseCases interface {
	// IsAvailable — площадка умеет покупки и продаёт хоть что-то
	IsAvailable() bool
	// GetRows — строки магазина по порядку конфигурации; пачки,
	// уже открытые во всех режимах, и пачки после покупки всех
	// уровней не продаются
	GetRows() []types.ShopRow
	// IsOnSale — в магазине есть товары вида kind
	IsOnSale(kind types.ProductKind) bool
	// RequestPurchase открывает окно оплаты товара площадки
	RequestPurchase(productID string)
	// PollPurchase — исход покупки; оплаченное сразу зачисляется
	PollPurchase() (types.PurchaseStatus, error)
	// Sync зачисляет покупки площадки: разовые — в инвентарь,
	// расходуемые — один раз, затем списывает их у площадки
	Sync() error
}
