package adapters

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

// TileStaticEntity представляет статический тайл с изображением
type TileStaticEntity struct {
	imageId string
}

// GetImageId возвращает ID изображения тайла
func (tse *TileStaticEntity) GetImageId() string {
	return tse.imageId
}

// TileAnimationEntity представляет анимированный тайл
type TileAnimationEntity struct {
	isPaused        bool
	isStopped       bool
	currentFrame    uint
	animationFrames types.AnimationData
}

// GetCurrentFrameImage возвращает изображение текущего кадра
func (tae *TileAnimationEntity) GetCurrentFrameImage() (string, error) {
	if len(tae.animationFrames) == 0 {
		return "", fmt.Errorf("no animation frames available")
	}

	if int(tae.currentFrame) >= len(tae.animationFrames) {
		return "", fmt.Errorf("current frame %d out of range", tae.currentFrame)
	}

	frame := tae.animationFrames[tae.currentFrame]
	return frame.Image, nil
}

// GetCurrentFrameDuration возвращает длительность текущего кадра
func (tae *TileAnimationEntity) GetCurrentFrameDuration() int {
	if len(tae.animationFrames) == 0 {
		return 0
	}

	if int(tae.currentFrame) >= len(tae.animationFrames) {
		return 0
	}

	return tae.animationFrames[tae.currentFrame].Duration
}

// NextFrame переключает на следующий кадр анимации
func (tae *TileAnimationEntity) NextFrame() {
	if tae.isPaused || tae.isStopped {
		return
	}

	if len(tae.animationFrames) == 0 {
		return
	}

	tae.currentFrame = (tae.currentFrame + 1) % uint(len(tae.animationFrames))
}

// Pause приостанавливает анимацию
func (tae *TileAnimationEntity) Pause() {
	tae.isPaused = true
}

// Resume возобновляет анимацию
func (tae *TileAnimationEntity) Resume() {
	tae.isPaused = false
}

// Stop останавливает анимацию
func (tae *TileAnimationEntity) Stop() {
	tae.isStopped = true
	tae.currentFrame = 0
}

// IsPaused возвращает true, если анимация приостановлена
func (tae *TileAnimationEntity) IsPaused() bool {
	return tae.isPaused
}

// IsStopped возвращает true, если анимация остановлена
func (tae *TileAnimationEntity) IsStopped() bool {
	return tae.isStopped
}

// TilesAdapter адаптер для работы с тайлами
type TilesAdapter struct {
	tilesRepository types.ITilesetRepository
}

// NewTilesAdapter создает новый адаптер тайлов
func NewTilesAdapter(tilesRepository types.ITilesetRepository) *TilesAdapter {
	return &TilesAdapter{
		tilesRepository: tilesRepository,
	}
}

// GetTilesetRepository возвращает репозиторий тайлсетов
func (ta *TilesAdapter) GetTilesetRepository() types.ITilesetRepository {
	return ta.tilesRepository
}

// GetTileStaticEntity создает статический тайл по ID изображения
func (ta *TilesAdapter) GetTileStaticEntity(id string) (types.IImageIdGetter, error) {
	// Проверяем, что изображение существует
	_, err := ta.tilesRepository.GetImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}

	return &TileStaticEntity{
		imageId: id,
	}, nil
}

// GetTileAnimationEntity создает анимированный тайл по ID анимации
func (ta *TilesAdapter) GetTileAnimationEntity(id string) (*TileAnimationEntity, error) {
	// Получаем данные анимации
	animationFrames, err := ta.tilesRepository.GetAnimationData(id)
	if err != nil {
		return nil, fmt.Errorf("animation '%s' not found: %w", id, err)
	}

	return &TileAnimationEntity{
		isPaused:        false,
		isStopped:       false,
		currentFrame:    0,
		animationFrames: animationFrames,
	}, nil
}

// CreateBlockEntity создает BlockEntity с TileStaticEntity
func (ta *TilesAdapter) CreateBlockEntity(blockType string, positionX, positionY float64) (*types.BlockEntity, error) {
	// Создаем TileStaticEntity для блока
	tileEntity, err := ta.GetTileStaticEntity(blockType)
	if err != nil {
		return nil, fmt.Errorf("failed to create tile entity for block type '%s': %w", blockType, err)
	}

	// Создаем BlockEntity
	blockEntity := &types.BlockEntity{
		ImageGetter: tileEntity,
		Data: &types.BlockData{
			Name:     types.BlockType(blockType),
			Position: types.Position{X: positionX, Y: positionY},
		},
		Properties: &types.BlockProperties{
			Collidable: true, // По умолчанию блоки коллизибельны
		},
		WorldPosition: types.Position{X: positionX, Y: positionY},
	}

	return blockEntity, nil
}
