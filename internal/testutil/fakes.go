// Package testutil содержит фейки тяжёлой инфраструктуры для тестов.
package testutil

import (
	"fmt"
	"image"

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

// FakeMenuInput реализует interfaces.IMenuInputAdapter: события
// кадра задаются полями, Reset очищает их перед следующим кадром
type FakeMenuInput struct {
	TouchActive bool
	Up, Down    bool
	Side        int
	Confirm     bool
	BackPressed bool
	Pause       bool
}

var _ interfaces.IMenuInputAdapter = (*FakeMenuInput)(nil)

func (f *FakeMenuInput) Update()                {}
func (f *FakeMenuInput) IsTouchActive() bool    { return f.TouchActive }
func (f *FakeMenuInput) Steps() (bool, bool)    { return f.Up, f.Down }
func (f *FakeMenuInput) SideStep() int          { return f.Side }
func (f *FakeMenuInput) Confirmed() bool        { return f.Confirm }
func (f *FakeMenuInput) Back() bool             { return f.BackPressed }
func (f *FakeMenuInput) PauseJustPressed() bool { return f.Pause }

// Reset — кадр без нажатий
func (f *FakeMenuInput) Reset() {
	*f = FakeMenuInput{TouchActive: f.TouchActive}
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
