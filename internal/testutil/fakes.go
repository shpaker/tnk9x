// Package testutil содержит фейки тяжёлой инфраструктуры для тестов.
package testutil

import (
	"fmt"
	"image"
	"sort"
	"strings"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

// FakeImageProvider реализует types.IImageProvider.
type FakeImageProvider struct {
	ImageID string
	Err     error
}

var _ types.IImageProvider = (*FakeImageProvider)(nil)

func (p *FakeImageProvider) GetImageID() (string, error) {
	return p.ImageID, p.Err
}

// FakeTileService реализует interfaces.ITileService и возвращает
// минимальные анимации без реальных тайлсетов.
type FakeTileService struct {
	Err     error    // если задана — CreateAnimationTileFromTileset падает
	Created []string // "tilesetType/id" всех созданных анимаций
}

var _ interfaces.ITileService = (*FakeTileService)(nil)

func (s *FakeTileService) GetTileAnimationFrames(
	id string,
) (types.AnimationData, error) {
	return types.AnimationData{{Image: id, Duration: 1}}, nil
}

func (s *FakeTileService) GetAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return types.AnimationConfig{Duration: 1, Frames: []string{id}}, nil
}

func (s *FakeTileService) CreateAnimationFromConfig(
	animationFrames types.AnimationData,
	config types.AnimationConfig,
) *image_providers.AnimationProvider {
	return image_providers.NewAnimationProvider(animationFrames)
}

// CreateAnimationTileFromTileset каждый раз возвращает новый
// AnimationProvider, чтобы работали пути спавна и взрывов.
func (s *FakeTileService) CreateAnimationTileFromTileset(
	tilesetType types.TilesetType,
	id string,
) (*image_providers.AnimationProvider, error) {
	s.Created = append(s.Created, string(tilesetType)+"/"+id)
	if s.Err != nil {
		return nil, s.Err
	}
	return image_providers.NewAnimationProvider(
		types.AnimationData{{Image: id, Duration: 1}},
	), nil
}

// FakeSoundPlayer реализует interfaces.ISoundPlayerAdapter и записывает
// вызовы в экспортируемые слайсы.
type FakeSoundPlayer struct {
	Played       []types.SoundID
	Looped       []types.SoundID
	Stopped      []types.SoundID
	StopAllCalls int
	UpdateCalls  int
	Volume       float64
}

var _ interfaces.ISoundPlayerAdapter = (*FakeSoundPlayer)(nil)

func (p *FakeSoundPlayer) Play(soundID types.SoundID) error {
	p.Played = append(p.Played, soundID)
	return nil
}

func (p *FakeSoundPlayer) PlayLoop(soundID types.SoundID) error {
	p.Looped = append(p.Looped, soundID)
	return nil
}

func (p *FakeSoundPlayer) Stop(soundID types.SoundID) {
	p.Stopped = append(p.Stopped, soundID)
}

func (p *FakeSoundPlayer) StopAll() {
	p.StopAllCalls++
}

func (p *FakeSoundPlayer) Update() {
	p.UpdateCalls++
}

func (p *FakeSoundPlayer) SetVolume(volume float64) {
	p.Volume = volume
}

// FakeTouchControls реализует interfaces.ITouchControlsAdapter:
// события кадра задаются полями по номеру игрока
type FakeTouchControls struct {
	TouchActive   bool
	Directions    [2]types.Direction
	HasDirection  [2]bool
	DirectionJust [2]bool
	FireJust      [2]bool
	PauseJust     bool
	Tap           types.Position
	TapJust       bool
}

var _ interfaces.ITouchControlsAdapter = (*FakeTouchControls)(nil)

func (f *FakeTouchControls) Update() {}

func (f *FakeTouchControls) IsTouchActive() bool { return f.TouchActive }

func (f *FakeTouchControls) DPadDirection(
	player types.PlayerTankNum,
) (types.Direction, bool) {
	return f.Directions[player], f.HasDirection[player]
}

func (f *FakeTouchControls) DPadJustPressed(
	player types.PlayerTankNum,
) (types.Direction, bool) {
	return f.Directions[player], f.DirectionJust[player]
}

func (f *FakeTouchControls) FireJustPressed(player types.PlayerTankNum) bool {
	return f.FireJust[player]
}

func (f *FakeTouchControls) PauseJustPressed() bool { return f.PauseJust }

func (f *FakeTouchControls) TapJustPressed() (types.Position, bool) {
	return f.Tap, f.TapJust
}

// GamePosition — логические координаты совпадают с координатами ebiten
func (f *FakeTouchControls) GamePosition(x, y int) (types.Position, bool) {
	return types.Position{X: float64(x), Y: float64(y)}, true
}

// FakeMenuInput реализует interfaces.IMenuInputAdapter: события
// кадра задаются полями, Reset очищает их перед следующим кадром
type FakeMenuInput struct {
	TouchActive bool
	Up, Down    bool
	Side        int
	Confirm     bool
	BackPressed bool
	Pause       bool
	TapPosition types.Position
	TapPressed  bool
	// Наведение мыши в этом кадре и защёлка «мышь замечена»
	PointPosition types.Position
	PointMoved    bool
	PointerActive bool
}

var _ interfaces.IMenuInputAdapter = (*FakeMenuInput)(nil)

func (f *FakeMenuInput) Update()                {}
func (f *FakeMenuInput) IsTouchActive() bool    { return f.TouchActive }
func (f *FakeMenuInput) Steps() (bool, bool)    { return f.Up, f.Down }
func (f *FakeMenuInput) SideStep() int          { return f.Side }
func (f *FakeMenuInput) Confirmed() bool        { return f.Confirm }
func (f *FakeMenuInput) Back() bool             { return f.BackPressed }
func (f *FakeMenuInput) PauseJustPressed() bool { return f.Pause }

func (f *FakeMenuInput) Tapped() (types.Position, bool) { return f.TapPosition, f.TapPressed }
func (f *FakeMenuInput) Pointed() (types.Position, bool) {
	return f.PointPosition, f.PointMoved
}
func (f *FakeMenuInput) IsPointerActive() bool { return f.PointerActive }

// Reset — кадр без нажатий
func (f *FakeMenuInput) Reset() {
	*f = FakeMenuInput{
		TouchActive:   f.TouchActive,
		PointerActive: f.PointerActive,
	}
}

// FakeTilesetRegistry реализует interfaces.ITilesetRepositoryRegistry:
// GetImageData возвращает пустую картинку 1x1, анимации — минимальный
// однокадровый набор. Requested хранит "тайлсет/id" всех запрошенных
// изображений. MissingIDs объявляет отдельные "тайлсет/id"
// несуществующими.
type FakeTilesetRegistry struct {
	Err        error       // если задана — все Get* падают
	Requested  []string    // "тайлсет/id" запрошенных изображений
	ImageIDs   []string    // ответ GetImageIDs для любого тайлсета
	MissingIDs []string    // "тайлсет/id", для которых Get* падают
	Image      image.Image // ответ GetImageData; nil — пустая картинка 1x1
}

var _ interfaces.ITilesetRepositoryRegistry = (*FakeTilesetRegistry)(nil)

func (r *FakeTilesetRegistry) missing(key string) bool {
	for _, missingID := range r.MissingIDs {
		if missingID == key {
			return true
		}
	}
	return false
}

func (r *FakeTilesetRegistry) GetImageData(
	tilesetType types.TilesetType,
	id string,
) (image.Image, error) {
	key := string(tilesetType) + "/" + id
	r.Requested = append(r.Requested, key)
	if r.Err != nil {
		return nil, r.Err
	}
	if r.missing(key) {
		return nil, fmt.Errorf("image '%s' not found", id)
	}
	if r.Image != nil {
		return r.Image, nil
	}
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}

func (r *FakeTilesetRegistry) GetAnimationData(
	tilesetType types.TilesetType,
	id string,
) (types.AnimationData, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if r.missing(string(tilesetType) + "/" + id) {
		return nil, fmt.Errorf("animation '%s' not found", id)
	}
	return types.AnimationData{{Image: "frame", Duration: 1}}, nil
}

func (r *FakeTilesetRegistry) GetAnimationConfig(
	tilesetType types.TilesetType,
	id string,
) (types.AnimationConfig, error) {
	if r.Err != nil {
		return types.AnimationConfig{}, r.Err
	}
	return types.AnimationConfig{Duration: 1, Frames: []string{"frame"}}, nil
}

func (r *FakeTilesetRegistry) GetImageIDs(
	tilesetType types.TilesetType,
) []string {
	return r.ImageIDs
}

// FakeVisualEffectsUseCases реализует interfaces.IVisualEffectsUseCases
// и записывает запрошенные события.
type FakeVisualEffectsUseCases struct {
	Events      []types.VisualEventEntity
	FlashLights []types.LightEntity
}

var _ interfaces.IVisualEffectsUseCases = (*FakeVisualEffectsUseCases)(nil)

func (f *FakeVisualEffectsUseCases) RequestEffect(
	event types.VisualEventEntity,
) {
	f.Events = append(f.Events, event)
}

func (f *FakeVisualEffectsUseCases) Update() {}

func (f *FakeVisualEffectsUseCases) GetParticles() []types.ParticleEntity {
	return nil
}

func (f *FakeVisualEffectsUseCases) GetFlashLights() []types.LightEntity {
	return f.FlashLights
}

func (f *FakeVisualEffectsUseCases) GetShakeOffset() types.Position {
	return types.Position{}
}

// Kinds возвращает виды записанных событий по порядку.
func (f *FakeVisualEffectsUseCases) Kinds() []types.VisualEventKind {
	kinds := make([]types.VisualEventKind, 0, len(f.Events))
	for _, event := range f.Events {
		kinds = append(kinds, event.Kind)
	}
	return kinds
}

// FakeTankCommonUseCases — общие use cases танков без логики;
// Frozen задаёт ответ IsFrozen
type FakeTankCommonUseCases struct {
	Frozen bool
}

var _ interfaces.ITankCommonUseCases = (*FakeTankCommonUseCases)(nil)

func (f *FakeTankCommonUseCases) Update(*types.TankEntity, float64) error {
	return nil
}

func (f *FakeTankCommonUseCases) UpdateAllTanks(float64) error { return nil }

func (f *FakeTankCommonUseCases) GetAllTanks() []*types.TankEntity {
	return nil
}

func (f *FakeTankCommonUseCases) GetAllPlayerTanks() []*types.TankEntity {
	return nil
}

func (f *FakeTankCommonUseCases) IsAnyPlayerTankMoving() bool { return false }

func (f *FakeTankCommonUseCases) LevelUp(*types.TankEntity) {}

func (f *FakeTankCommonUseCases) LevelDown(*types.TankEntity) {}

func (f *FakeTankCommonUseCases) SetMaxLevel(*types.TankEntity) {}

func (f *FakeTankCommonUseCases) IsFrozen(*types.TankEntity) bool {
	return f.Frozen
}

// FakePlatform реализует interfaces.IPlatformAdapter: приостановка
// задаётся полями, вызовы записываются
type FakePlatform struct {
	Suspended         bool
	JustSuspended     bool
	ReadyCalls        int
	IntermissionCalls int
	RatingCalls       int
	// Gameplay — значения SetGameplayActive по порядку
	Gameplay []bool
	// Language — язык площадки
	Language string
}

var _ interfaces.IPlatformAdapter = (*FakePlatform)(nil)

func (f *FakePlatform) Update()               {}
func (f *FakePlatform) IsSuspended() bool     { return f.Suspended }
func (f *FakePlatform) IsJustSuspended() bool { return f.JustSuspended }
func (f *FakePlatform) Ready()                { f.ReadyCalls++ }
func (f *FakePlatform) RequestIntermission()  { f.IntermissionCalls++ }
func (f *FakePlatform) RequestRating()        { f.RatingCalls++ }
func (f *FakePlatform) GetLanguage() string   { return f.Language }

func (f *FakePlatform) SetGameplayActive(active bool) {
	f.Gameplay = append(f.Gameplay, active)
}

// LastGameplay — последнее переданное значение разметки геймплея
func (f *FakePlatform) LastGameplay() (bool, bool) {
	if len(f.Gameplay) == 0 {
		return false, false
	}
	return f.Gameplay[len(f.Gameplay)-1], true
}

// FakeWindow реализует interfaces.IWindowAdapter: полный экран
// и видимость курсора хранятся в полях
type FakeWindow struct {
	Fullscreen   bool
	CursorHidden bool
}

var _ interfaces.IWindowAdapter = (*FakeWindow)(nil)

func (f *FakeWindow) IsFullscreen() bool { return f.Fullscreen }

func (f *FakeWindow) SetFullscreen(
	fullscreen bool,
) {
	f.Fullscreen = fullscreen
}

func (f *FakeWindow) SetCursorVisible(
	visible bool,
) {
	f.CursorHidden = !visible
}

// FakeReward реализует interfaces.IRewardAdapter: итоговый Status
// отдаётся PollReward один раз, как у настоящей площадки
type FakeReward struct {
	Available bool
	Requests  int
	Status    types.RewardStatus
}

var _ interfaces.IRewardAdapter = (*FakeReward)(nil)

func (f *FakeReward) IsRewardAvailable() bool { return f.Available }
func (f *FakeReward) RequestReward()          { f.Requests++ }

func (f *FakeReward) PollReward() types.RewardStatus {
	status := f.Status
	if status == types.RewardStatusGranted ||
		status == types.RewardStatusDenied {
		f.Status = types.RewardStatusNone
	}
	return status
}

// FakePurchase реализует interfaces.IPurchaseAdapter: каталог
// и покупки задаются полями, итоговый Status отдаётся PollPurchase
// один раз, списанные покупки уходят из Purchases
type FakePurchase struct {
	Available bool
	Catalog   []types.ProductOffer
	Purchases []types.Purchase
	Status    types.PurchaseStatus
	// Requests — ID товаров запросов покупки по порядку
	Requests []string
	// Consumed — токены списанных покупок по порядку
	Consumed []string
}

var _ interfaces.IPurchaseAdapter = (*FakePurchase)(nil)

func (f *FakePurchase) IsPurchaseAvailable() bool        { return f.Available }
func (f *FakePurchase) GetCatalog() []types.ProductOffer { return f.Catalog }
func (f *FakePurchase) GetPurchases() []types.Purchase   { return f.Purchases }

func (f *FakePurchase) RequestPurchase(
	productID string,
) {
	f.Requests = append(f.Requests, productID)
}

func (f *FakePurchase) PollPurchase() types.PurchaseStatus {
	status := f.Status
	if status == types.PurchaseStatusGranted ||
		status == types.PurchaseStatusDenied {
		f.Status = types.PurchaseStatusNone
	}
	return status
}

func (f *FakePurchase) ConsumePurchase(token string) {
	f.Consumed = append(f.Consumed, token)
	kept := f.Purchases[:0]
	for _, purchase := range f.Purchases {
		if purchase.Token != token {
			kept = append(kept, purchase)
		}
	}
	f.Purchases = kept
}

// FakeInventory реализует interfaces.IInventoryUseCases: жетоны
// и отключённая реклама — в полях
type FakeInventory struct {
	NoAds  bool
	Tokens uint
}

var _ interfaces.IInventoryUseCases = (*FakeInventory)(nil)

func (f *FakeInventory) HasNoAds() bool  { return f.NoAds }
func (f *FakeInventory) GetTokens() uint { return f.Tokens }

func (f *FakeInventory) SpendToken() (bool, error) {
	if f.Tokens == 0 {
		return false, nil
	}
	f.Tokens--
	return true, nil
}

// MemoryInventoryRepository реализует interfaces.IInventoryRepository
// в памяти: считает сохранения, Err — ошибка сохранения
type MemoryInventoryRepository struct {
	Saves int
	Err   error
}

var _ interfaces.IInventoryRepository = (*MemoryInventoryRepository)(nil)

func (r *MemoryInventoryRepository) GetInventory() (*types.InventoryEntity, error) {
	return types.NewInventoryEntity(), nil
}

func (r *MemoryInventoryRepository) SaveInventory(
	*types.InventoryEntity,
) error {
	r.Saves++
	return r.Err
}

// FakeTexts реализует interfaces.ITextsAdapter: текст — сам ключ,
// параметры дописываются через пробел в порядке имён; языки
// конфигурации — Languages, название языка — его код в верхнем регистре
type FakeTexts struct {
	Language  types.Language
	Languages []types.Language
}

var _ interfaces.ITextsAdapter = (*FakeTexts)(nil)

func (f *FakeTexts) GetLanguage() types.Language { return f.Language }

func (f *FakeTexts) SetLanguage(language types.Language) error {
	for _, supported := range f.Languages {
		if supported == language {
			f.Language = language
			return nil
		}
	}
	return fmt.Errorf("unsupported language %q", language)
}

func (f *FakeTexts) Get(key types.TextKey) string { return string(key) }

func (f *FakeTexts) Format(key types.TextKey, args types.TextArgs) string {
	names := make([]string, 0, len(args))
	for name := range args {
		names = append(names, name)
	}
	sort.Strings(names)
	text := string(key)
	for _, name := range names {
		text += fmt.Sprintf(" %v", args[name])
	}
	return text
}

func (f *FakeTexts) Plural(
	key types.TextKey,
	count int,
	args types.TextArgs,
) string {
	return f.Format(key, args) + fmt.Sprintf(" %d", count)
}

func (f *FakeTexts) GetOr(_ types.TextKey, fallback string) string {
	return fallback
}

func (f *FakeTexts) GetLanguageName(language types.Language) string {
	return strings.ToUpper(string(language))
}
