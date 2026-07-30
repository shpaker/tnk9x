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

// FakeTilesetRegistry реализует interfaces.ITilesetRepositoryRegistry:
// GetImageData возвращает пустую картинку 1x1, анимации — минимальный
// однокадровый набор. Requested хранит "тайлсет/id" всех запрошенных
// изображений. MissingIDs объявляет отдельные "тайлсет/id"
// несуществующими.
type FakeTilesetRegistry struct {
	Err        error    // если задана — все Get* падают
	Requested  []string // "тайлсет/id" запрошенных изображений
	ImageIDs   []string // ответ GetImageIDs для любого тайлсета
	MissingIDs []string // "тайлсет/id", для которых Get* падают
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
