// Package testutil содержит фейки тяжёлой инфраструктуры для тестов.
package testutil

import (
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
	tilesetType string,
	id string,
) (*image_providers.AnimationProvider, error) {
	s.Created = append(s.Created, tilesetType+"/"+id)
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

func (p *FakeSoundPlayer) Stop(soundID types.SoundID) error {
	p.Stopped = append(p.Stopped, soundID)
	return nil
}

func (p *FakeSoundPlayer) StopAll() {
	p.StopAllCalls++
}

func (p *FakeSoundPlayer) Update() error {
	p.UpdateCalls++
	return nil
}

// FakeTilesetRegistry реализует interfaces.ITilesetRepositoryRegistry:
// каждый Get*Image возвращает провайдер с запрошенным id, GetImageData —
// пустую картинку 1x1. Requested хранит id всех запрошенных изображений.
type FakeTilesetRegistry struct {
	Err       error    // если задана — все Get*Image падают
	Requested []string // id запрошенных изображений
}

var _ interfaces.ITilesetRepositoryRegistry = (*FakeTilesetRegistry)(nil)

func (r *FakeTilesetRegistry) image(id string) (types.IImageProvider, error) {
	r.Requested = append(r.Requested, id)
	if r.Err != nil {
		return nil, r.Err
	}
	return &FakeImageProvider{ImageID: id}, nil
}

func (r *FakeTilesetRegistry) animationData() (types.AnimationData, error) {
	return types.AnimationData{{Image: "frame", Duration: 1}}, nil
}

func (r *FakeTilesetRegistry) animationConfig() (types.AnimationConfig, error) {
	return types.AnimationConfig{Duration: 1, Frames: []string{"frame"}}, nil
}

func (r *FakeTilesetRegistry) GetBlocksImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetBlocksAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetBlocksAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetPlayerImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetPlayerAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetPlayerAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetEnemyImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetEnemyAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetEnemyAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetBulletImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetBulletAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetBulletAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetSpawnerImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetSpawnerAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetSpawnerAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetExplosionTankImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetExplosionTankAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetExplosionTankAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetBulletExplosionImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetBulletExplosionAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetBulletExplosionAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetHQImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetHQAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetHQAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetBonusesImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetBonusesAnimationData(
	id string,
) (types.AnimationData, error) {
	return r.animationData()
}

func (r *FakeTilesetRegistry) GetBonusesAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	return r.animationConfig()
}

func (r *FakeTilesetRegistry) GetHUDImage(
	id string,
) (types.IImageProvider, error) {
	return r.image(id)
}

func (r *FakeTilesetRegistry) GetImageData(
	tilesetType string,
	id string,
) (image.Image, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
}
